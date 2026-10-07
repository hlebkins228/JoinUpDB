//go:build allure

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
	"JoinUp/internal/testutil"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/mock"
)

type stubReviewEventRepo struct {
	service.EventRepo
	event *models.Event
}

func (r *stubReviewEventRepo) GetEvent(context.Context, int) (*models.Event, error) {
	return r.event, nil
}

func reviewEventFixture() *models.Event {
	event := testEventModel()
	event.ID = 7
	event.EventTime = time.Now().UTC().Add(-24 * time.Hour)
	return event
}

type classicReviewFixture struct {
	ctx        context.Context
	cancel     context.CancelFunc
	connection *testutil.ReviewConnection
	reviews    *postgres.ReviewRepo
	svc        service.ReviewService
}

func newClassicReviewFixture() classicReviewFixture {
	ctx, cancel := context.WithCancel(context.Background())
	connection := testutil.NewReviewConnection()
	reviews := postgres.NewReviewRepo(connection)
	uow := postgres.NewUOW(connection)
	events := &stubReviewEventRepo{event: reviewEventFixture()}
	return classicReviewFixture{
		ctx: ctx, cancel: cancel, connection: connection, reviews: &reviews,
		svc: service.NewReviewService(&uow, &reviews, events),
	}
}

type mockReviewRepo struct {
	mock.Mock
}

func (m *mockReviewRepo) Create(ctx context.Context, review *models.Review) (*models.Review, error) {
	args := m.Called(ctx, review)
	created, _ := args.Get(0).(*models.Review)
	return created, args.Error(1)
}

func (m *mockReviewRepo) ListByEvent(ctx context.Context, eventID int) ([]*models.Review, error) {
	args := m.Called(ctx, eventID)
	reviews, _ := args.Get(0).([]*models.Review)
	return reviews, args.Error(1)
}

func (m *mockReviewRepo) ExistsByEventAndAuthor(ctx context.Context, eventID, authorID int) (bool, error) {
	args := m.Called(ctx, eventID, authorID)
	return args.Bool(0), args.Error(1)
}

type reviewTxKey struct{}

type londonReviewFixture struct {
	ctx     context.Context
	cancel  context.CancelFunc
	txCtx   context.Context
	events  *mockEventRepo
	reviews *mockReviewRepo
	uow     *mockUOW
	svc     service.ReviewService
}

func newLondonReviewFixture() londonReviewFixture {
	ctx, cancel := context.WithCancel(context.Background())
	txCtx := context.WithValue(ctx, reviewTxKey{}, "review transaction")
	events, reviews, uow := new(mockEventRepo), new(mockReviewRepo), new(mockUOW)
	return londonReviewFixture{
		ctx:     ctx,
		cancel:  cancel,
		txCtx:   txCtx,
		events:  events,
		reviews: reviews,
		uow:     uow,
		svc:     service.NewReviewService(uow, reviews, events),
	}
}

func (f londonReviewFixture) assertExpectations(pt provider.T) {
	f.events.AssertExpectations(pt)
	f.reviews.AssertExpectations(pt)
	f.uow.AssertExpectations(pt)
}

func TestAllureReviewServiceClassic(t *testing.T) {
	r := runner.NewRunner(t, "ReviewService — классический подход")
	var f classicReviewFixture
	r.BeforeAll(func(pt provider.T) {
		pt.WithNewStep("BeforeAll — создание зависимостей", func(_ provider.StepCtx) {
			f = newClassicReviewFixture()
			pt.Require().False(f.connection.Closed())
		})
	})
	r.AfterAll(func(pt provider.T) {
		pt.WithNewStep("AfterAll — освобождение зависимостей", func(_ provider.StepCtx) {
			f.connection.Close()
			f.cancel()
			pt.Assert().True(f.connection.Closed())
			pt.Assert().ErrorIs(f.ctx.Err(), context.Canceled)
			f = classicReviewFixture{}
		})
	})
	r.BeforeEach(func(pt provider.T) {
		pt.WithNewStep("BeforeEach — очистка состояния теста", func(_ provider.StepCtx) {
			f.connection.Reset()
			pt.Epic("JoinUp — ЛР1")
			pt.Feature("ReviewService")
			pt.Tags("classic", "unit", "reviews", "fake", "stub")
		})
	})
	r.NewTest("Создание: 2000 Unicode-символов и удаление пробелов", func(pt provider.T) {
		text := strings.Repeat("я", 2000)
		input := testutil.NewReviewBuilder().WithText(" \n" + text + "\t ").Build()

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Require().NoError(err)
		pt.Require().NotNil(got)
		pt.Assert().Equal(1, got.ID)
		pt.Assert().Equal(input.EventID, got.EventID)
		pt.Assert().Equal(input.AuthorID, got.AuthorID)
		pt.Assert().Equal(input.Rating, got.Rating)
		pt.Assert().Equal(text, got.Text)
		pt.Assert().False(got.CreatedAt.IsZero())
		pt.Assert().Equal(time.UTC, got.CreatedAt.Location())
		stored, err := f.reviews.ListByEvent(f.ctx, input.EventID)
		pt.Require().NoError(err)
		pt.Require().Len(stored, 1)
		pt.Assert().Equal(got.Text, stored[0].Text)
		pt.Assert().Equal(got.ID, stored[0].ID)
	})
	r.NewTest("Повторный отзыв не меняет хранилище", func(pt provider.T) {
		input := testutil.NewReviewBuilder().Build()
		first, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)
		pt.Require().NoError(err)

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, "Другой отзыв")

		pt.Assert().ErrorIs(err, exceptions.ErrAlreadyExists)
		pt.Assert().Nil(got)
		stored, err := f.reviews.ListByEvent(f.ctx, input.EventID)
		pt.Require().NoError(err)
		pt.Require().Len(stored, 1)
		pt.Assert().Equal(first.ID, stored[0].ID)
		pt.Assert().Equal(input.Text, stored[0].Text)
	})
	r.NewTest("2001 Unicode-символ: отзыв не сохраняется", func(pt provider.T) {
		input := testutil.NewReviewBuilder().WithText(strings.Repeat("я", 2001)).Build()

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrInvalidPayload)
		pt.Assert().Nil(got)
		stored, err := f.reviews.ListByEvent(f.ctx, input.EventID)
		pt.Require().NoError(err)
		pt.Assert().Empty(stored)
	})
	r.RunTests()
}

func TestAllureReviewServiceLondon(t *testing.T) {
	r := runner.NewRunner(t, "ReviewService — лондонский подход")
	var f londonReviewFixture
	r.BeforeAll(func(pt provider.T) {
		pt.WithNewStep("BeforeAll — создание зависимостей", func(_ provider.StepCtx) {
			f = newLondonReviewFixture()
		})
	})
	r.AfterAll(func(pt provider.T) {
		pt.WithNewStep("AfterAll — освобождение зависимостей", func(_ provider.StepCtx) {
			f.cancel()
			pt.Assert().ErrorIs(f.ctx.Err(), context.Canceled)
			f = londonReviewFixture{}
		})
	})
	r.AfterEach(func(pt provider.T) {
		pt.WithNewStep("AfterEach — проверка ожиданий mock", func(_ provider.StepCtx) {
			f.assertExpectations(pt)
		})
	})
	r.BeforeEach(func(pt provider.T) {
		pt.WithNewStep("BeforeEach — очистка состояния теста", func(_ provider.StepCtx) {
			f.events.Mock = mock.Mock{}
			f.reviews.Mock = mock.Mock{}
			f.uow.Mock = mock.Mock{}
			f.events.Test(pt)
			f.reviews.Test(pt)
			f.uow.Test(pt)
			f.uow.On("BeginTx", f.ctx).Return(f.txCtx, nil).Once()
			f.uow.On("Rollback", f.txCtx).Return(nil).Once()
			pt.Epic("JoinUp — ЛР1")
			pt.Feature("ReviewService")
			pt.Tags("london", "unit", "reviews", "mock")
		})
	})
	r.NewTest("Создание: нормализация текста и commit", func(pt provider.T) {
		event := reviewEventFixture()
		f.events.On("GetEvent", f.txCtx, event.ID).Return(event, nil).Once()
		input := testutil.NewReviewBuilder().WithRating(1).WithText("  Хорошо!\n").Build()
		created := testutil.NewReviewBuilder().WithRating(1).WithText("Хорошо!").Build()
		created.ID = 42
		f.reviews.On("ExistsByEventAndAuthor", f.txCtx, input.EventID, input.AuthorID).Return(false, nil).Once()
		f.reviews.On("Create", f.txCtx, mock.MatchedBy(func(rv *models.Review) bool {
			return rv.EventID == input.EventID && rv.AuthorID == input.AuthorID &&
				rv.Rating == 1 && rv.Text == "Хорошо!" && !rv.CreatedAt.IsZero() && rv.CreatedAt.Location() == time.UTC
		})).Return(created, nil).Once()
		f.uow.On("Commit", f.txCtx).Return(nil).Once()

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Require().NoError(err)
		pt.Require().NotNil(got)
		pt.Assert().Equal(created.ID, got.ID)
		pt.Assert().Equal(created.Text, got.Text)
		pt.Assert().Equal(created.Rating, got.Rating)
		pt.Assert().Equal(created.Author.Name, got.Author.Name)
	})
	r.NewTest("Будущее событие: без сохранения и commit", func(pt provider.T) {
		event := reviewEventFixture()
		event.EventTime = time.Now().UTC().Add(24 * time.Hour)
		f.events.On("GetEvent", f.txCtx, event.ID).Return(event, nil).Once()
		input := testutil.NewReviewBuilder().Build()

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, exceptions.ErrEventNotPassed)
		pt.Assert().Nil(got)
		f.reviews.AssertNotCalled(pt, "Create", mock.Anything, mock.Anything)
		f.uow.AssertNotCalled(pt, "Commit", mock.Anything)
	})
	r.NewTest("Ошибка сохранения: rollback без commit", func(pt provider.T) {
		event := reviewEventFixture()
		f.events.On("GetEvent", f.txCtx, event.ID).Return(event, nil).Once()
		input := testutil.NewReviewBuilder().Build()
		dbErr := errors.New("database unavailable")
		f.reviews.On("ExistsByEventAndAuthor", f.txCtx, input.EventID, input.AuthorID).Return(false, nil).Once()
		f.reviews.On("Create", f.txCtx, mock.Anything).Return(nil, dbErr).Once()

		got, err := f.svc.CreateReview(f.ctx, input.EventID, input.AuthorID, input.Rating, input.Text)

		pt.Assert().ErrorIs(err, dbErr)
		pt.Assert().Nil(got)
		f.uow.AssertNotCalled(pt, "Commit", mock.Anything)
	})
	r.RunTests()
}
