package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var defaultCategories = []mongoCategory{
	{ID: 1, Name: "спорт"},
	{ID: 2, Name: "музыка"},
	{ID: 3, Name: "игры"},
	{ID: 4, Name: "арт"},
	{ID: 5, Name: "прогулки"},
	{ID: 6, Name: "учёба"},
	{ID: 7, Name: "еда"},
	{ID: 8, Name: "tech"},
	{ID: 9, Name: "путешествия"},
	{ID: 10, Name: "фото"},
	{ID: 11, Name: "кино"},
}

func EnsureCategories(ctx context.Context, client *mongo.Client, dbName string) error {
	coll := client.Database(dbName).Collection("categories")

	for _, cat := range defaultCategories {
		filter := bson.M{"_id": cat.ID}
		update := bson.M{"$setOnInsert": bson.M{"name": cat.Name}}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
	}

	return nil
}
