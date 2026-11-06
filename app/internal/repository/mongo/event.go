package mongo

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const eventsCollection = "events"

type EventRepo struct {
	Pool
}

func NewEventRepo(client *mongo.Client, dbName string) EventRepo {
	return EventRepo{Pool: Pool{client: client, dbName: dbName}}
}

func (r *EventRepo) CreateEvent(ctx context.Context, event *models.Event) (int, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	id, err := nextID(ctx, r.collection("counters"), eventsCollection)
	if err != nil {
		return 0, err
	}
	event.ID = id

	// populate creator info from users collection
	var creator mongoUser
	if err := r.collection(usersCollection).FindOne(ctx, bson.M{"_id": event.CreatorID}).Decode(&creator); err == nil {
		event.Creator = models.EventMember{
			ID:       creator.ID,
			Name:     creator.Name,
			AvatarID: creator.AvatarID,
		}
	}

	_, err = r.collection(eventsCollection).InsertOne(ctx, fromModelEvent(event))
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *EventRepo) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var me mongoEvent
	err = r.collection(eventsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&me)
	if err == mongo.ErrNoDocuments {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	event := me.toModel()
	r.ensureCreator(ctx, event)
	return event, nil
}

func (r *EventRepo) ensureCreator(ctx context.Context, event *models.Event) {
	if event.Creator.ID != 0 {
		return
	}
	var creator mongoUser
	if err := r.collection(usersCollection).FindOne(ctx, bson.M{"_id": event.CreatorID}).Decode(&creator); err == nil {
		event.Creator = models.EventMember{
			ID:       creator.ID,
			Name:     creator.Name,
			AvatarID: creator.AvatarID,
		}
	}
}

func (r *EventRepo) UpdateEvent(ctx context.Context, event *models.Event) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	fields := bson.M{
		"name":              event.Name,
		"description":       event.Desc,
		"updated_at":        event.UpdatedAt,
		"event_date":        event.EventTime,
		"telegram_chat_url": event.TelegramChatURL,
		"city":              event.City,
		"image_id":          event.ImageID,
		"location":          fromModelLocation(event.Location),
	}
	if event.Categories != nil {
		fields["categories"] = modelCategoriesToMongo(event.Categories)
	}

	result, err := r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": event.ID, "deleted": false},
		bson.M{"$set": fields},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *EventRepo) DeleteEvent(ctx context.Context, id int) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	var me mongoEvent
	err = r.collection(eventsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&me)
	if err == mongo.ErrNoDocuments {
		return exceptions.ErrNotFound
	}
	if err != nil {
		return err
	}
	if me.Deleted {
		return exceptions.ErrAlreadyDeleted
	}

	_, err = r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": id, "deleted": false},
		bson.M{"$set": bson.M{"deleted": true, "updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *EventRepo) JoinEvent(ctx context.Context, eventID int, userID int) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	// get user info to embed in members
	var user mongoUser
	err = r.collection(usersCollection).FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return exceptions.ErrNotExists
	}
	if err != nil {
		return err
	}

	member := fromModelEventMember(models.EventMember{
		ID:       user.ID,
		Name:     user.Name,
		AvatarID: user.AvatarID,
	})

	result, err := r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": eventID, "deleted": false, "members.id": bson.M{"$ne": userID}},
		bson.M{"$push": bson.M{"members": member}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		// check if event exists at all
		var me mongoEvent
		err = r.collection(eventsCollection).FindOne(ctx, bson.M{"_id": eventID}).Decode(&me)
		if err == mongo.ErrNoDocuments || me.Deleted {
			return exceptions.ErrNotFound
		}
		return exceptions.ErrAlreadyExists
	}

	return nil
}

func (r *EventRepo) LeaveEvent(ctx context.Context, eventID int, userID int) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	result, err := r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": eventID, "deleted": false, "members.id": userID},
		bson.M{"$pull": bson.M{"members": bson.M{"id": userID}}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		var me mongoEvent
		err = r.collection(eventsCollection).FindOne(ctx, bson.M{"_id": eventID}).Decode(&me)
		if err == mongo.ErrNoDocuments || me.Deleted {
			return exceptions.ErrNotFound
		}
		return exceptions.ErrNotExists
	}

	return nil
}

func (r *EventRepo) IsParticipant(ctx context.Context, eventID, userID int) (bool, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	count, err := r.collection(eventsCollection).CountDocuments(ctx,
		bson.M{"_id": eventID, "members.id": userID},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *EventRepo) UpdateEventImage(ctx context.Context, eventID int, imageID int, updatedAt time.Time) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	result, err := r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": eventID, "deleted": false},
		bson.M{"$set": bson.M{"image_id": imageID, "updated_at": updatedAt}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return exceptions.ErrNotFound
	}

	return nil
}

func (r *EventRepo) AddEventCategory(ctx context.Context, eventID int, categoryID int) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	// check event exists
	var me mongoEvent
	err = r.collection(eventsCollection).FindOne(ctx, bson.M{"_id": eventID, "deleted": false}).Decode(&me)
	if err == mongo.ErrNoDocuments {
		return exceptions.ErrNotFound
	}
	if err != nil {
		return err
	}

	// check category exists
	var mc mongoCategory
	err = r.collection("categories").FindOne(ctx, bson.M{"_id": categoryID}).Decode(&mc)
	if err == mongo.ErrNoDocuments {
		return exceptions.ErrNotExists
	}
	if err != nil {
		return err
	}

	// check if already added
	for _, c := range me.Categories {
		if c.ID == categoryID {
			return exceptions.ErrAlreadyExists
		}
	}

	_, err = r.collection(eventsCollection).UpdateOne(ctx,
		bson.M{"_id": eventID},
		bson.M{"$push": bson.M{"categories": mc}},
	)
	return err
}

func (r *EventRepo) SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	query := bson.M{"deleted": false}

	if filter.Name != nil && *filter.Name != "" {
		query["name"] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.City != nil && *filter.City != "" {
		query["city"] = *filter.City
	}
	if filter.EventFrom != nil || filter.EventTo != nil {
		dateFilter := bson.M{}
		if filter.EventFrom != nil {
			dateFilter["$gte"] = *filter.EventFrom
		}
		if filter.EventTo != nil {
			dateFilter["$lte"] = *filter.EventTo
		}
		query["event_date"] = dateFilter
	}
	if len(filter.CategoryIDs) > 0 {
		query["categories._id"] = bson.M{"$all": filter.CategoryIDs}
	}

	opts := options.Find().SetSort(bson.D{bson.E{Key: "event_date", Value: 1}, bson.E{Key: "_id", Value: 1}})

	cursor, err := r.collection(eventsCollection).Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mongoEvents := make([]*mongoEvent, 0)
	if err := cursor.All(ctx, &mongoEvents); err != nil {
		return nil, err
	}

	events := make([]*models.Event, 0, len(mongoEvents))
	for _, me := range mongoEvents {
		e := me.toModel()
		r.ensureCreator(ctx, e)
		events = append(events, e)
	}

	return events, nil
}

func (r *EventRepo) ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{bson.E{Key: "event_date", Value: 1}, bson.E{Key: "_id", Value: 1}})
	cursor, err := r.collection(eventsCollection).Find(ctx,
		bson.M{"creator_id": userID, "deleted": false},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mongoEvents := make([]*mongoEvent, 0)
	if err := cursor.All(ctx, &mongoEvents); err != nil {
		return nil, err
	}

	events := make([]*models.Event, 0, len(mongoEvents))
	for _, me := range mongoEvents {
		e := me.toModel()
		r.ensureCreator(ctx, e)
		events = append(events, e)
	}

	return events, nil
}

func (r *EventRepo) ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{bson.E{Key: "event_date", Value: 1}, bson.E{Key: "_id", Value: 1}})
	cursor, err := r.collection(eventsCollection).Find(ctx,
		bson.M{"members.id": userID, "deleted": false},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mongoEvents := make([]*mongoEvent, 0)
	if err := cursor.All(ctx, &mongoEvents); err != nil {
		return nil, err
	}

	events := make([]*models.Event, 0, len(mongoEvents))
	for _, me := range mongoEvents {
		e := me.toModel()
		r.ensureCreator(ctx, e)
		events = append(events, e)
	}

	return events, nil
}

func (r *EventRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{bson.E{Key: "name", Value: 1}, bson.E{Key: "_id", Value: 1}})
	cursor, err := r.collection("categories").Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mongoCats := make([]*mongoCategory, 0)
	if err := cursor.All(ctx, &mongoCats); err != nil {
		return nil, err
	}

	categories := make([]*models.Category, 0, len(mongoCats))
	for _, mc := range mongoCats {
		cat := mc.toModel()
		categories = append(categories, &cat)
	}

	return categories, nil
}
