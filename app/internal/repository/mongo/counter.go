package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func nextID(ctx context.Context, coll *mongo.Collection, counterName string) (int, error) {
	filter := bson.M{"_id": counterName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		return 0, err
	}
	return result.Seq, nil
}
