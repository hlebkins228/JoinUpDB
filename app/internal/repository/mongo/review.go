package mongo

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const reviewsCollection = "reviews"

type ReviewRepo struct {
	Pool
}

func NewReviewRepo(client *mongo.Client, dbName string) ReviewRepo {
	return ReviewRepo{Pool: Pool{client: client, dbName: dbName}}
}

func (r *ReviewRepo) Create(ctx context.Context, rv *models.Review) (*models.Review, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	id, err := nextID(ctx, r.collection("counters"), reviewsCollection)
	if err != nil {
		return nil, err
	}
	rv.ID = id

	// fetch author info to embed
	var author mongoReviewAuthor
	err = r.collection(usersCollection).FindOne(ctx, bson.M{"_id": rv.AuthorID}).Decode(&author)
	if err == nil {
		rv.Author = author.toModel()
	}

	_, err = r.collection(reviewsCollection).InsertOne(ctx, fromModelReview(rv))
	if mongo.IsDuplicateKeyError(err) {
		return nil, fmt.Errorf("%w: %v", exceptions.ErrAlreadyExists, err)
	}
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (r *ReviewRepo) ListByEvent(ctx context.Context, eventID int) ([]*models.Review, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{bson.E{Key: "created_at", Value: -1}, bson.E{Key: "_id", Value: -1}})
	cursor, err := r.collection(reviewsCollection).Find(ctx, bson.M{"event_id": eventID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mongoReviews := make([]*mongoReview, 0)
	if err := cursor.All(ctx, &mongoReviews); err != nil {
		return nil, err
	}

	reviews := make([]*models.Review, 0, len(mongoReviews))
	for _, mr := range mongoReviews {
		reviews = append(reviews, mr.toModel())
	}

	return reviews, nil
}

func (r *ReviewRepo) ExistsByEventAndAuthor(ctx context.Context, eventID, authorID int) (bool, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	err = r.collection(reviewsCollection).FindOne(ctx,
		bson.M{"event_id": eventID, "author_id": authorID},
	).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// EnsureReviewIndexes creates required indexes for the reviews collection.
func EnsureReviewIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	coll := client.Database(dbName).Collection(reviewsCollection)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				bson.E{Key: "event_id", Value: 1},
				bson.E{Key: "author_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}
