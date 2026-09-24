package postgres

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepo struct {
	Pool
}

type eventScanner interface {
	Scan(dest ...any) error
}

const eventSelectFields = `
	e.id, e.name, e.description,
	e.created_at at time zone 'utc',
	e.updated_at at time zone 'utc',
	e.event_date at time zone 'utc',
	e.telegram_chat_url, e.city, e.creator_id, e.image_id,
	l.name, l.longitude, l.latitude, l.address, e.deleted
`

func NewEventRepo(pool *pgxpool.Pool) EventRepo {
	return EventRepo{Pool: Pool{pool: pool}}
}

func scanEvent(row eventScanner) (*models.Event, error) {
	event := models.Event{}
	err := row.Scan(
		&event.ID,
		&event.Name,
		&event.Desc,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.EventTime,
		&event.TelegramChatURL,
		&event.City,
		&event.CreatorID,
		&event.ImageID,
		&event.Location.Name,
		&event.Location.Longitude,
		&event.Location.Latitude,
		&event.Location.Address,
		&event.Deleted,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *EventRepo) CreateEvent(ctx context.Context, event *models.Event) (int, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	const locationInsertQuery = `
		insert into location (name, longitude, latitude, address)
		values ($1, $2, $3, $4)
		returning id
	`

	const eventInsertQuery = `
		insert into event (name, description, created_at, updated_at, event_date, telegram_chat_url, city, creator_id, location_id, image_id)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		returning id
	`

	var locationID int
	err = run.QueryRow(ctx, locationInsertQuery, event.Location.Name,
		event.Location.Longitude, event.Location.Latitude, event.Location.Address).Scan(&locationID)
	if err != nil {
		return 0, err
	}

	var eventID int
	err = run.QueryRow(ctx, eventInsertQuery, event.Name, event.Desc,
		event.CreatedAt, event.UpdatedAt, event.EventTime, event.TelegramChatURL,
		event.City, event.CreatorID, locationID, event.ImageID).Scan(&eventID)
	if err != nil {
		return 0, err
	}

	return eventID, nil
}

func (r *EventRepo) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
	select ` + eventSelectFields + `
	from event e
	join location l on e.location_id = l.id
	where e.id = $1`

	event, err := scanEvent(run.QueryRow(ctx, query, id))
	if err != nil {
		return nil, err
	}
	if err = r.loadEventMembers(ctx, run, event); err != nil {
		return nil, err
	}
	if err = r.loadEventCreators(ctx, run, event); err != nil {
		return nil, err
	}
	if err = r.loadEventCategories(ctx, run, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepo) UpdateEvent(ctx context.Context, event *models.Event) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	const locationIDQuery = `
		select location_id
		from event
		where id = $1 and deleted = false
	`

	const locationUpdateQuery = `
		update location
		set name = $1, longitude = $2, latitude = $3, address = $4
		where id = $5
	`

	const eventUpdateQuery = `
		update event
		set name = $1,
		    description = $2,
		    updated_at = $3,
		    event_date = $4,
		    telegram_chat_url = $5,
		    city = $6,
		    image_id = $7
		where id = $8 and deleted = false
	`

	var locationID int
	if err = run.QueryRow(ctx, locationIDQuery, event.ID).Scan(&locationID); err != nil {
		return err
	}

	if _, err = run.Exec(ctx, locationUpdateQuery,
		event.Location.Name, event.Location.Longitude, event.Location.Latitude, event.Location.Address, locationID,
	); err != nil {
		return err
	}

	tag, err := run.Exec(ctx, eventUpdateQuery,
		event.Name, event.Desc, event.UpdatedAt, event.EventTime, event.TelegramChatURL,
		event.City, event.ImageID, event.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *EventRepo) DeleteEvent(ctx context.Context, id int) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	var deleted bool
	err = run.QueryRow(ctx, `select deleted from event where id = $1`, id).Scan(&deleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.ErrNotFound
		}
		return err
	}
	if deleted {
		return exceptions.ErrAlreadyDeleted
	}

	now := time.Now().UTC()
	const query = `
		update event
		set deleted = true, updated_at = $2
		where id = $1 and deleted = false
	`

	_, err = run.Exec(ctx, query, id, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) JoinEvent(ctx context.Context, eventID int, userID int) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	var eventExists bool
	err = run.QueryRow(ctx, `select exists(select 1 from event where id = $1 and deleted = false)`, eventID).Scan(&eventExists)
	if err != nil {
		return err
	}
	if !eventExists {
		return exceptions.ErrNotFound
	}

	var alreadyMember bool
	err = run.QueryRow(ctx, `select exists(select 1 from member where event_id = $1 and user_id = $2)`, eventID, userID).Scan(&alreadyMember)
	if err != nil {
		return err
	}
	if alreadyMember {
		return exceptions.ErrAlreadyExists
	}

	const query = `
		insert into member (user_id, event_id, role)
		values ($1, $2, 'member')
	`

	_, err = run.Exec(ctx, query, userID, eventID)
	return err
}

func (r *EventRepo) LeaveEvent(ctx context.Context, eventID int, userID int) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	var eventExists bool
	err = run.QueryRow(ctx, `select exists(select 1 from event where id = $1 and deleted = false)`, eventID).Scan(&eventExists)
	if err != nil {
		return err
	}
	if !eventExists {
		return exceptions.ErrNotFound
	}

	tag, err := run.Exec(ctx, `delete from member where event_id = $1 and user_id = $2`, eventID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return exceptions.ErrNotExists
	}

	return nil
}

func (r *EventRepo) IsParticipant(ctx context.Context, eventID, userID int) (bool, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	const query = `select exists(select 1 from member where event_id = $1 and user_id = $2)`

	var ok bool
	err = run.QueryRow(ctx, query, eventID, userID).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *EventRepo) UpdateEventImage(ctx context.Context, eventID int, imageID int, updatedAt time.Time) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	const query = `
		update event
		set image_id = $1, updated_at = $2
		where id = $3 and deleted = false
	`

	tag, err := run.Exec(ctx, query, imageID, updatedAt, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return exceptions.ErrNotFound
	}

	return nil
}

func (r *EventRepo) AddEventCategory(ctx context.Context, eventID int, categoryID int) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	var eventExists bool
	err = run.QueryRow(ctx, `select exists(select 1 from event where id = $1 and deleted = false)`, eventID).Scan(&eventExists)
	if err != nil {
		return err
	}
	if !eventExists {
		return exceptions.ErrNotFound
	}

	var categoryExists bool
	err = run.QueryRow(ctx, `select exists(select 1 from category where id = $1)`, categoryID).Scan(&categoryExists)
	if err != nil {
		return err
	}
	if !categoryExists {
		return exceptions.ErrNotExists
	}

	var alreadyAdded bool
	err = run.QueryRow(ctx, `
		select exists(
			select 1
			from event_category
			where event_id = $1 and category_id = $2
		)
	`, eventID, categoryID).Scan(&alreadyAdded)
	if err != nil {
		return err
	}
	if alreadyAdded {
		return exceptions.ErrAlreadyExists
	}

	const query = `
		insert into event_category (event_id, category_id)
		values ($1, $2)
	`

	_, err = run.Exec(ctx, query, eventID, categoryID)
	return err
}

func (r *EventRepo) SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]int32, 0, len(filter.CategoryIDs))
	for _, id := range filter.CategoryIDs {
		categoryIDs = append(categoryIDs, int32(id))
	}

	const query = `
		select ` + eventSelectFields + `
		from event e
		join location l on e.location_id = l.id
		where e.deleted = false
		  and ($1::text is null or e.name ilike '%' || $1 || '%')
		  and ($2::timestamp is null or e.event_date >= $2)
		  and ($3::timestamp is null or e.event_date <= $3)
		  and ($4::text is null or e.city = $4)
		  and (
			cardinality($5::int[]) = 0
			or e.id in (
				select ec.event_id
				from event_category ec
				where ec.category_id = any($5::int[])
				group by ec.event_id
				having count(distinct ec.category_id) = cardinality($5::int[])
			)
		  )
		order by e.event_date, e.id
	`

	rows, err := run.Query(ctx, query,
		filter.Name,
		filter.EventFrom,
		filter.EventTo,
		filter.City,
		categoryIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*models.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err = r.loadEventsMembers(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCreators(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCategories(ctx, run, events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select ` + eventSelectFields + `
		from event e
		join location l on e.location_id = l.id
		where e.creator_id = $1 and e.deleted = false
		order by e.event_date, e.id
	`

	rows, err := run.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*models.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err = r.loadEventsMembers(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCreators(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCategories(ctx, run, events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select ` + eventSelectFields + `
		from member m
		join event e on e.id = m.event_id
		join location l on e.location_id = l.id
		where m.user_id = $1 and e.deleted = false
		order by e.event_date, e.id
	`

	rows, err := run.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*models.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err = r.loadEventsMembers(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCreators(ctx, run, events); err != nil {
		return nil, err
	}
	if err = r.loadEventsCategories(ctx, run, events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select id, name
		from category
		order by name, id
	`

	rows, err := run.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*models.Category, 0)
	for rows.Next() {
		category := models.Category{}
		if err = rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return categories, nil
}

func (r *EventRepo) loadEventMembers(ctx context.Context, run sqlRun, event *models.Event) error {
	return r.loadEventsMembers(ctx, run, []*models.Event{event})
}

func (r *EventRepo) loadEventCreators(ctx context.Context, run sqlRun, event *models.Event) error {
	return r.loadEventsCreators(ctx, run, []*models.Event{event})
}

func (r *EventRepo) loadEventCategories(ctx context.Context, run sqlRun, event *models.Event) error {
	return r.loadEventsCategories(ctx, run, []*models.Event{event})
}

func (r *EventRepo) loadEventsCreators(ctx context.Context, run sqlRun, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	creatorIDs := make([]int32, 0, len(events))
	byCreatorID := make(map[int][]*models.Event, len(events))
	for _, event := range events {
		creatorIDs = append(creatorIDs, int32(event.CreatorID))
		byCreatorID[event.CreatorID] = append(byCreatorID[event.CreatorID], event)
	}

	const query = `
		select id, name, avatar_id
		from users
		where id = any($1::int[])
	`

	rows, err := run.Query(ctx, query, creatorIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var creator models.EventMember
		if err = rows.Scan(&creator.ID, &creator.Name, &creator.AvatarID); err != nil {
			return err
		}
		for _, event := range byCreatorID[creator.ID] {
			event.Creator = creator
		}
	}

	return rows.Err()
}

func (r *EventRepo) loadEventsMembers(ctx context.Context, run sqlRun, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	eventIDs := make([]int32, 0, len(events))
	byID := make(map[int]*models.Event, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, int32(event.ID))
		event.Members = []models.EventMember{}
		byID[event.ID] = event
	}

	const query = `
		select m.event_id, u.id, u.name, u.avatar_id
		from member m
		join users u on u.id = m.user_id
		where m.event_id = any($1::int[])
		order by u.name, u.id
	`

	rows, err := run.Query(ctx, query, eventIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var eventID int
		var member models.EventMember
		if err = rows.Scan(&eventID, &member.ID, &member.Name, &member.AvatarID); err != nil {
			return err
		}
		if event, ok := byID[eventID]; ok {
			event.Members = append(event.Members, member)
		}
	}

	return rows.Err()
}

func (r *EventRepo) loadEventsCategories(ctx context.Context, run sqlRun, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	eventIDs := make([]int32, 0, len(events))
	byID := make(map[int]*models.Event, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, int32(event.ID))
		event.Categories = []models.Category{}
		byID[event.ID] = event
	}

	const query = `
		select ec.event_id, c.id, c.name
		from event_category ec
		join category c on c.id = ec.category_id
		where ec.event_id = any($1::int[])
		order by c.name, c.id
	`

	rows, err := run.Query(ctx, query, eventIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var eventID int
		var category models.Category
		if err = rows.Scan(&eventID, &category.ID, &category.Name); err != nil {
			return err
		}
		if event, ok := byID[eventID]; ok {
			event.Categories = append(event.Categories, category)
		}
	}

	return rows.Err()
}
