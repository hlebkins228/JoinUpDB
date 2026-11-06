package postgres

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return UserRepo{Pool: Pool{pool: pool}}
}

func (r *UserRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	const query = `select exists(select 1 from users where id = $1)`
	var ok bool
	err = run.QueryRow(ctx, query, id).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *UserRepo) AddUser(ctx context.Context, user *models.User) (int, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	const query = `
		insert into users (name, age, login, password, created_at, city, telegram_login, avatar_id)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict(login) do nothing
		returning id
	`

	var id int = -1
	err = run.QueryRow(ctx, query,
		user.Name,
		user.Age,
		user.Login,
		user.Password,
		user.CreatedAt,
		user.City,
		user.TelegramLogin,
		user.AvatarID,
	).Scan(&id)

	if err != nil {
		if id == -1 {
			return 0, fmt.Errorf("%w: %v", exceptions.ErrAlreadyExists, err)
		}
		return 0, err
	}

	return id, nil
}

func (r *UserRepo) GetUser(ctx context.Context, id int) (*models.User, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select
			u.id, u.name, u.age, u.login, u.created_at, u.city, u.telegram_login, u.avatar_id, u.role,
			(select count(*) from event e where e.creator_id = u.id and e.deleted = false),
			(select count(*) from member m join event e on e.id = m.event_id where m.user_id = u.id and e.deleted = false),
			(select count(*) from review r join event e on e.id = r.event_id where e.creator_id = u.id and e.deleted = false),
			(select avg(r.rating)::float8 from review r join event e on e.id = r.event_id where e.creator_id = u.id and e.deleted = false)
		from users u
		where u.id = $1
	`

	user := models.User{}
	err = run.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Age,
		&user.Login,
		&user.CreatedAt,
		&user.City,
		&user.TelegramLogin,
		&user.AvatarID,
		&user.Role,
		&user.CreatedEventsCount,
		&user.JoinedEventsCount,
		&user.OrganizerReviewsCount,
		&user.OrganizerRating,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select id, name, age, login, password, created_at, city, telegram_login, avatar_id, role
		from users
		where login = $1
	`

	user := models.User{}
	err = run.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Name,
		&user.Age,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.City,
		&user.TelegramLogin,
		&user.AvatarID,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	const query = `
		update users
		set name = $1,
		    age = $2,
		    city = $3,
		    telegram_login = $4,
		    avatar_id = $5
		where id = $6
	`

	tag, err := run.Exec(ctx, query,
		user.Name,
		user.Age,
		user.City,
		user.TelegramLogin,
		user.AvatarID,
		user.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: user id=%d", exceptions.ErrNotFound, user.ID)
	}

	return nil
}
