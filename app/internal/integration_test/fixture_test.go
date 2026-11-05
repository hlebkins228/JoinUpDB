//go:build integration

package integration_test

import (
	"context"
	"os"

	"JoinUp/internal/repository/postgres"
	"JoinUp/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func reviewFixture(pt provider.T) (context.Context, *pgxpool.Pool, service.ReviewService) {
	ctx := context.Background()
	url := os.Getenv("TEST_DATABASE_URL")
	pt.Require().NotEmpty(url, "TEST_DATABASE_URL must point to the Docker test database")
	pool, err := pgxpool.New(ctx, url)
	pt.Require().NoError(err)
	pt.Cleanup(pool.Close)
	pt.Require().Equal("joinup_test", pool.Config().ConnConfig.Database)
	reset, err := os.ReadFile("../../tests/reset.sql")
	pt.Require().NoError(err)
	pt.Cleanup(func() {
		_, err := pool.Exec(ctx, string(reset))
		pt.Assert().NoError(err)
	})
	_, err = pool.Exec(ctx, string(reset))
	pt.Require().NoError(err)
	seed, err := os.ReadFile("../../tests/fixtures.sql")
	pt.Require().NoError(err)
	_, err = pool.Exec(ctx, string(seed))
	pt.Require().NoError(err)
	uow := postgres.NewUOW(pool)
	reviews := postgres.NewReviewRepo(pool)
	events := postgres.NewEventRepo(pool)
	return ctx, pool, service.NewReviewService(&uow, &reviews, &events)
}
