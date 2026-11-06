package mongo

import (
	"JoinUp/internal/models"
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const imagesCollection = "images"

type ImageRepo struct {
	Pool
}

func NewImageRepo(client *mongo.Client, dbName string) ImageRepo {
	return ImageRepo{Pool: Pool{client: client, dbName: dbName}}
}

func (r *ImageRepo) AddImage(ctx context.Context, img *models.Image) (int, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	id, err := nextID(ctx, r.collection("counters"), imagesCollection)
	if err != nil {
		return 0, err
	}

	doc := fromModelImage(img)
	doc.ID = id

	_, err = r.collection(imagesCollection).InsertOne(ctx, doc)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ImageRepo) GetImage(ctx context.Context, id int) (*models.Image, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var mi mongoImage
	err = r.collection(imagesCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&mi)
	if err == mongo.ErrNoDocuments {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return mi.toModel(), nil
}

func (r *ImageRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	ctx, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	err = r.collection(imagesCollection).FindOne(ctx, bson.M{"_id": id}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
