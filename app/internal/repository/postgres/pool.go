package postgres

import (
	"context"
)

type Pool struct {
	pool sqlRun
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
