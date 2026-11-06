package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Pool struct {
	client *mongo.Client
	dbName string
}

func (p *Pool) collection(name string) *mongo.Collection {
	return p.client.Database(p.dbName).Collection(name)
}

func (p *Pool) runFromCtx(ctx context.Context) (context.Context, error) {
	sess, err := TxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return ctx, nil
	}
	return mongo.NewSessionContext(ctx, sess), nil
}
