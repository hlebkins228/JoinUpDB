//go:build integration

package integration_test

import (
	"testing"

	"JoinUp/internal/exceptions"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestIntegrationReviewFuture(t *testing.T) {
	runner.Run(t, "Future event review is rejected", func(pt provider.T) {
		ctx, pool, svc := reviewFixture(pt)

		_, err := svc.CreateReview(ctx, 2, 2, 5, "Too early")

		pt.Assert().ErrorIs(err, exceptions.ErrEventNotPassed)
		var count int
		err = pool.QueryRow(ctx, "select count(*) from review").Scan(&count)
		pt.Require().NoError(err)
		pt.Assert().Zero(count)
	})
}
