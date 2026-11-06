//go:build integration

package integration_test

import (
	"testing"

	"JoinUp/internal/exceptions"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestIntegrationReviewDuplicate(t *testing.T) {
	runner.Run(t, "Duplicate review is rejected", func(pt provider.T) {
		ctx, pool, svc := reviewFixture(pt)
		_, err := svc.CreateReview(ctx, 1, 2, 5, "Original review")
		pt.Require().NoError(err)

		_, err = svc.CreateReview(ctx, 1, 2, 1, "Duplicate review")

		pt.Assert().ErrorIs(err, exceptions.ErrAlreadyExists)
		var count int
		var text string
		err = pool.QueryRow(ctx, "select count(*), min(text) from review where event_id = 1").Scan(&count, &text)
		pt.Require().NoError(err)
		pt.Assert().Equal(1, count)
		pt.Assert().Equal("Original review", text)
	})
}
