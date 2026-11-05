package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"JoinUp/internal/repository/postgres"
	"JoinUp/internal/service"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/mock"
)

type reviewBuilder struct {
	review models.Review
}

func (b reviewBuilder) WithRating(rating int) reviewBuilder {
	b.review.Rating = rating
	return b
}

func (b reviewBuilder) WithText(text string) reviewBuilder {
	b.review.Text = text
	return b
}

func (b reviewBuilder) Build() *models.Review {
	review := b.review
	return &review
}

type stubReviewEventRepo struct {
	service.EventRepo
	event *models.Event
}

func (r *stubReviewEventRepo) GetEvent(context.Context, int) (*models.Event, error) {
	return r.event, nil
}

type mockEventRepo struct {
	service.EventRepo
	mock.Mock
}

func (m *mockEventRepo) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	args := m.Called(ctx, id)
	event, _ := args.Get(0).(*models.Event)
	return event, args.Error(1)
}

type mockUOW struct {
	mock.Mock
}

func (m *mockUOW) BeginTx(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	return args.Get(0).(context.Context), args.Error(1)
}

func (m *mockUOW) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockUOW) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type mockReviewRepo struct {
	service.ReviewRepo
	mock.Mock
}

func (m *mockReviewRepo) Create(ctx context.Context, review *models.Review) (*models.Review, error) {
	args := m.Called(ctx, review)
	created, _ := args.Get(0).(*models.Review)
	return created, args.Error(1)
}

func (m *mockReviewRepo) ExistsByEventAndAuthor(ctx context.Context, eventID, authorID int) (bool, error) {
	args := m.Called(ctx, eventID, authorID)
	return args.Bool(0), args.Error(1)
}

type reviewTxKey struct{}

func TestAllureReviewServiceClassic(t *testing.T) {
	r := runner.NewRunner(t, "ReviewService Classic")
	ctx := context.Background()
	date := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	longText := strings.Repeat("t", 2000)
	var (
		pool  pgxmock.PgxPoolIface
		svc   service.ReviewService
		input *models.Review
		event *models.Event
	)
	r.BeforeEach(func(pt provider.T) {
		builder := reviewBuilder{review: models.Review{EventID: 7, AuthorID: 20}}
		input = builder.WithRating(5).WithText("Great meetup").Build()
		event = &models.Event{ID: 7, CreatorID: 10, EventTime: date}
		var err error
		pool, err = pgxmock.NewPool()
		pt.Require().NoError(err)
		reviews := postgres.NewReviewRepo(pool)
		uow := postgres.NewUOW(pool)
		svc = service.NewReviewService(&uow, &reviews, &stubReviewEventRepo{event: event})
	})
	r.AfterEach(func(pt provider.T) {
		pool.ExpectClose()
		pool.Close()
		pt.Assert().NoError(pool.ExpectationsWereMet())
	})
	r.NewTest("Create trims 2000 Unicode characters", func(pt provider.T) {
		input.Text = " \n" + longText + "\t "
		pool.ExpectBegin()
		pool.ExpectQuery("select exists").WithArgs(input.EventID, input.AuthorID).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
		pool.ExpectQuery("insert into review").
			WithArgs(input.EventID, input.AuthorID, input.Rating, longText, pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(42, date))
		pool.ExpectCommit()
		pool.ExpectRollback()

		got, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Require().NoError(err)
		pt.Assert().Equal(42, got.ID)
		pt.Assert().Equal(longText, got.Text)
	})
	r.NewTest("Duplicate stops before insert", func(pt provider.T) {
		pool.ExpectBegin()
		pool.ExpectQuery("select exists").WithArgs(input.EventID, input.AuthorID).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
		pool.ExpectRollback()

		_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrAlreadyExists)
	})
	r.NewTest("Invalid text is rejected", func(pt provider.T) {
		for _, text := range []string{" \n\t ", longText + "é"} {
			_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, text)

			pt.Assert().ErrorIs(err, exceptions.ErrInvalidPayload)
		}
	})
	r.NewTest("Out-of-range rating is rejected", func(pt provider.T) {
		for _, rating := range []int{0, 6} {
			_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, rating, input.Text)

			pt.Assert().ErrorIs(err, exceptions.ErrInvalidPayload)
		}
	})
	r.NewTest("Organizer cannot review own event", func(pt provider.T) {
		pool.ExpectBegin()
		pool.ExpectRollback()

		_, err := svc.CreateReview(ctx, input.EventID, event.CreatorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrOrganizerCannotReview)
	})
	r.RunTests()
}

func TestAllureReviewServiceLondon(t *testing.T) {
	r := runner.NewRunner(t, "ReviewService London")
	ctx := context.Background()
	txCtx := context.WithValue(ctx, reviewTxKey{}, true)
	date := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	future := time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC)
	saveErr, commitErr := errors.New("save failed"), errors.New("commit failed")
	var (
		events         *mockEventRepo
		reviews        *mockReviewRepo
		uow            *mockUOW
		svc            service.ReviewService
		event          *models.Event
		input, created *models.Review
	)
	r.BeforeEach(func(pt provider.T) {
		builder := reviewBuilder{review: models.Review{EventID: 7, AuthorID: 20, CreatedAt: date}}
		input = builder.WithRating(5).WithText("Great meetup").Build()
		created = builder.WithRating(1).WithText("Good!").Build()
		created.ID = 42
		event = &models.Event{ID: 7, CreatorID: 10, EventTime: date}
		events, reviews, uow = new(mockEventRepo), new(mockReviewRepo), new(mockUOW)
		events.Test(pt)
		reviews.Test(pt)
		uow.Test(pt)
		events.On("GetEvent", txCtx, event.ID).Return(event, nil).Once()
		uow.On("BeginTx", ctx).Return(txCtx, nil).Once()
		uow.On("Rollback", txCtx).Return(nil).Once()
		svc = service.NewReviewService(uow, reviews, events)
	})
	r.AfterEach(func(pt provider.T) {
		events.AssertExpectations(pt)
		reviews.AssertExpectations(pt)
		uow.AssertExpectations(pt)
	})
	r.NewTest("Create passes normalized data and commits", func(pt provider.T) {
		input.Rating, input.Text = 1, "  Good!\n"
		check := reviews.On("ExistsByEventAndAuthor", txCtx, input.EventID, input.AuthorID).
			Return(false, nil).Once()
		save := reviews.On("Create", txCtx, mock.MatchedBy(func(review *models.Review) bool {
			return review.EventID == input.EventID && review.AuthorID == input.AuthorID &&
				review.Rating == created.Rating && review.Text == created.Text
		})).Return(created, nil).Once().NotBefore(check)
		uow.On("Commit", txCtx).Return(nil).Once().NotBefore(save)

		got, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Require().NoError(err)
		pt.Assert().Equal(created.ID, got.ID)
		pt.Assert().Equal(created.Text, got.Text)
	})
	r.NewTest("Future event stops before persistence", func(pt provider.T) {
		event.EventTime = future

		_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrEventNotPassed)
	})
	r.NewTest("Deleted event stops before persistence", func(pt provider.T) {
		event.Deleted = true

		_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrNotFound)
	})
	r.NewTest("Save failure rolls back without commit", func(pt provider.T) {
		reviews.On("ExistsByEventAndAuthor", txCtx, input.EventID, input.AuthorID).
			Return(false, nil).Once()
		reviews.On("Create", txCtx, mock.Anything).Return(nil, saveErr).Once()

		_, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, saveErr)
	})
	r.NewTest("Commit failure returns no review and rolls back", func(pt provider.T) {
		reviews.On("ExistsByEventAndAuthor", txCtx, input.EventID, input.AuthorID).
			Return(false, nil).Once()
		save := reviews.On("Create", txCtx, mock.Anything).Return(created, nil).Once()
		uow.On("Commit", txCtx).Return(commitErr).Once().NotBefore(save)

		got, err := svc.CreateReview(ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, commitErr)
		pt.Assert().Nil(got)
	})
	r.RunTests()
}
