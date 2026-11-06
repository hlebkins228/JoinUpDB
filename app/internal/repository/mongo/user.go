package mongo

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"database/sql"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const usersCollection = "users"

type UserRepo struct {
	Pool
}

func NewUserRepo(client *mongo.Client, dbName string) UserRepo {
	return UserRepo{Pool: Pool{client: client, dbName: dbName}}
}

func (r *UserRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	coll := r.collection(usersCollection)
	err = coll.FindOne(ctx, bson.M{"_id": id}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *UserRepo) AddUser(ctx context.Context, user *models.User) (int, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	id, err := nextID(ctx, r.collection("counters"), usersCollection)
	if err != nil {
		return 0, err
	}
	user.ID = id

	_, err = r.collection(usersCollection).InsertOne(ctx, fromModelUser(user))
	if mongo.IsDuplicateKeyError(err) {
		return 0, fmt.Errorf("%w: %v", exceptions.ErrAlreadyExists, err)
	}
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserRepo) GetUser(ctx context.Context, id int) (*models.User, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var mu mongoUser
	err = r.collection(usersCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&mu)
	if err == mongo.ErrNoDocuments {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	user := mu.toModel()

	// compute aggregate fields
	if createdCount, err := r.collection("events").CountDocuments(ctx,
		bson.M{"creator_id": id, "deleted": false}); err == nil {
		user.CreatedEventsCount = int(createdCount)
	}
	if joinedCount, err := r.collection("events").CountDocuments(ctx,
		bson.M{"members.id": id, "deleted": false}); err == nil {
		user.JoinedEventsCount = int(joinedCount)
	}

	// organizer reviews: find events created by this user, then count reviews for those events
	var creatorEventIDs []int
	cursor, err := r.collection("events").Find(ctx,
		bson.M{"creator_id": id, "deleted": false},
		options.Find().SetProjection(bson.M{"_id": 1}),
	)
	if err == nil {
		var docs []struct {
			ID int `bson:"_id"`
		}
		if err := cursor.All(ctx, &docs); err == nil {
			for _, d := range docs {
				creatorEventIDs = append(creatorEventIDs, d.ID)
			}
		}
	}
	if len(creatorEventIDs) > 0 {
		if reviewCount, err := r.collection("reviews").CountDocuments(ctx,
			bson.M{"event_id": bson.M{"$in": creatorEventIDs}}); err == nil {
			user.OrganizerReviewsCount = int(reviewCount)
		}

		pipeline := mongo.Pipeline{
			bson.D{bson.E{Key: "$match", Value: bson.M{"event_id": bson.M{"$in": creatorEventIDs}}}},
			bson.D{bson.E{Key: "$group", Value: bson.M{"_id": nil, "avg": bson.M{"$avg": "$rating"}}}},
		}
		cursor, err := r.collection("reviews").Aggregate(ctx, pipeline)
		if err == nil {
			defer cursor.Close(ctx)
			var result []struct {
				Avg *float64 `bson:"avg"`
			}
			if cursor.All(ctx, &result) == nil && len(result) > 0 {
				user.OrganizerRating = result[0].Avg
			}
		}
	}

	return user, nil
}

func (r *UserRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var mu mongoUser
	err = r.collection(usersCollection).FindOne(ctx, bson.M{"login": login}).Decode(&mu)
	if err == mongo.ErrNoDocuments {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return mu.toModel(), nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return err
	}

	result, err := r.collection(usersCollection).UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"name":           user.Name,
			"age":            user.Age,
			"city":           user.City,
			"telegram_login": user.TelegramLogin,
			"avatar_id":      user.AvatarID,
		}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("%w: user id=%d", exceptions.ErrNotFound, user.ID)
	}

	return nil
}

// EnsureIndexes creates required indexes for the users collection.
func EnsureUserIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	coll := client.Database(dbName).Collection(usersCollection)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{bson.E{Key: "login", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}
