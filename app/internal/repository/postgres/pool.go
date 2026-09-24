package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	pool *pgxpool.Pool
}

func (p *Pool) runFromCtx(ctx context.Context) (sqlRun, error) {
	var run sqlRun
	run, err := TxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if run == nil {
		run = p.pool
	}

	return run, nil
}
