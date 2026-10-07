//go:build allure

package postgres

import (
	"context"
	"testing"

	"JoinUp/internal/exceptions"
	"JoinUp/internal/testutil"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestAllureReviewRepo(t *testing.T) {
	r := runner.NewRunner(t, "ReviewRepo — PostgreSQL")
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		connection *testutil.ReviewConnection
		repo       ReviewRepo
	)
	r.BeforeAll(func(pt provider.T) {
		pt.WithNewStep("BeforeAll — создание зависимостей", func(_ provider.StepCtx) {
			ctx, cancel = context.WithCancel(context.Background())
			connection = testutil.NewReviewConnection()
			repo = NewReviewRepo(connection)
			pt.Require().False(connection.Closed())
		})
	})
	r.AfterAll(func(pt provider.T) {
		pt.WithNewStep("AfterAll — освобождение зависимостей", func(_ provider.StepCtx) {
			connection.Close()
			cancel()
			pt.Assert().True(connection.Closed())
			pt.Assert().ErrorIs(ctx.Err(), context.Canceled)
			repo = ReviewRepo{}
			connection, ctx, cancel = nil, nil, nil
		})
	})
	r.BeforeEach(func(pt provider.T) {
		pt.WithNewStep("BeforeEach — очистка состояния теста", func(_ provider.StepCtx) {
			connection.Reset()
			pt.Epic("JoinUp — ЛР1")
			pt.Feature("ReviewRepo")
			pt.Tags("unit", "reviews", "postgres", "fake-connection")
		})
	})
	r.NewTest("Создание: параметры SQL и чтение RETURNING", func(pt provider.T) {
		input := testutil.NewReviewBuilder().Build()
		wantArgs := []any{input.EventID, input.AuthorID, input.Rating, input.Text, input.CreatedAt}

		got, err := repo.Create(ctx, input)

		pt.Require().NoError(err)
		pt.Require().NotNil(got)
		pt.Assert().Equal(1, got.ID)
		pt.Assert().Equal(input.CreatedAt, got.CreatedAt)
		pt.Assert().Equal(wantArgs, connection.Args)
		pt.Assert().Contains(connection.QueryText, "insert into review")
		stored, err := repo.ListByEvent(ctx, input.EventID)
		pt.Require().NoError(err)
		pt.Require().Len(stored, 1)
		pt.Assert().Equal(got.ID, stored[0].ID)
		pt.Assert().Equal(input.Text, stored[0].Text)
		pt.Assert().Equal(input.Rating, stored[0].Rating)
	})
	r.NewTest("Нарушение уникальности: ErrAlreadyExists", func(pt provider.T) {
		input := testutil.NewReviewBuilder().Build()
		_, err := repo.Create(ctx, input)
		pt.Require().NoError(err)

		got, err := repo.Create(ctx, testutil.NewReviewBuilder().Build())

		pt.Assert().ErrorIs(err, exceptions.ErrAlreadyExists)
		pt.Assert().Nil(got)
		stored, err := repo.ListByEvent(ctx, input.EventID)
		pt.Require().NoError(err)
		pt.Assert().Len(stored, 1)
	})
	r.NewTest("Проверка существования по событию и автору", func(pt provider.T) {
		input := testutil.NewReviewBuilder().Build()
		_, err := repo.Create(ctx, input)
		pt.Require().NoError(err)

		got, err := repo.ExistsByEventAndAuthor(ctx, input.EventID, input.AuthorID)

		pt.Require().NoError(err)
		pt.Assert().True(got)
		pt.Assert().Equal([]any{input.EventID, input.AuthorID}, connection.Args)
	})
	r.RunTests()
}
