package mongo

import (
	"JoinUp/internal/exceptions"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UOW struct {
	client *mongo.Client
}

type uowKey struct{}

func NewUOW(client *mongo.Client) UOW {
	return UOW{client: client}
}

func (t *UOW) BeginTx(ctx context.Context) (context.Context, error) {
	sess, err := t.client.StartSession()
	if err != nil {
		return ctx, err
	}
	if err := sess.StartTransaction(options.Transaction()); err != nil {
		sess.EndSession(ctx)
		return ctx, err
	}
	return context.WithValue(ctx, uowKey{}, sess), nil
}

func (t *UOW) Commit(ctx context.Context) error {
	sess, err := t.txFromCtx(ctx)
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	return sess.CommitTransaction(ctx)
}

func (t *UOW) Rollback(ctx context.Context) error {
	sess, err := t.txFromCtx(ctx)
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	return sess.AbortTransaction(ctx)
}

func (t *UOW) txFromCtx(ctx context.Context) (*mongo.Session, error) {
	v := ctx.Value(uowKey{})
	if v == nil {
		return nil, exceptions.ErrNoTxObj
	}

	sess, ok := v.(*mongo.Session)
	if !ok {
		return nil, exceptions.ErrTxType
	}

	return sess, nil
}

func TxFromCtx(ctx context.Context) (*mongo.Session, error) {
	v := ctx.Value(uowKey{})
	if v == nil {
		return nil, nil
	}

	sess, ok := v.(*mongo.Session)
	if !ok {
		return nil, exceptions.ErrTxType
	}

	return sess, nil
}
