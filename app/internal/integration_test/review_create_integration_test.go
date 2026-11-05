//go:build integration

package integration_test

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestIntegrationReviewCreate(t *testing.T) {
	runner.Run(t, "Review creation and reading", func(pt provider.T) {
		ctx, _, svc := reviewFixture(pt)

		created, err := svc.CreateReview(ctx, 1, 2, 5, "  Отличная встреча!\n")

		pt.Require().NoError(err)
		listed, err := svc.ListReviews(ctx, 1)
		pt.Require().NoError(err)
		pt.Require().Len(listed.Reviews, 1)
		pt.Assert().Equal(created.ID, listed.Reviews[0].ID)
		pt.Assert().Equal("Отличная встреча!", listed.Reviews[0].Text)
		pt.Assert().Equal(5, listed.Reviews[0].Rating)
		pt.Assert().Equal("Reviewer", listed.Reviews[0].Author.Name)
	})
}
