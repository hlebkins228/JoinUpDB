package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"JoinUp/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEventRepo struct {
	mock.Mock
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, event *models.Event) (int, error) {
	args := m.Called(ctx, event)
	return args.Int(0), args.Error(1)
}

func (m *mockEventRepo) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	args := m.Called(ctx, id)
	ev, _ := args.Get(0).(*models.Event)
	return ev, args.Error(1)
}

func (m *mockEventRepo) UpdateEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepo) DeleteEvent(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEventRepo) JoinEvent(ctx context.Context, eventID int, userID int) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

func (m *mockEventRepo) LeaveEvent(ctx context.Context, eventID int, userID int) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

func (m *mockEventRepo) IsParticipant(ctx context.Context, eventID int, userID int) (bool, error) {
	args := m.Called(ctx, eventID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventRepo) UpdateEventImage(ctx context.Context, eventID int, imageID int, updatedAt time.Time) error {
	args := m.Called(ctx, eventID, imageID, updatedAt)
	return args.Error(0)
}

func (m *mockEventRepo) AddEventCategory(ctx context.Context, eventID int, categoryID int) error {
	args := m.Called(ctx, eventID, categoryID)
	return args.Error(0)
}

func (m *mockEventRepo) SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error) {
	args := m.Called(ctx, filter)
	events, _ := args.Get(0).([]*models.Event)
	return events, args.Error(1)
}

func (m *mockEventRepo) ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	args := m.Called(ctx, userID)
	events, _ := args.Get(0).([]*models.Event)
	return events, args.Error(1)
}

func (m *mockEventRepo) ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	args := m.Called(ctx, userID)
	events, _ := args.Get(0).([]*models.Event)
	return events, args.Error(1)
}

func (m *mockEventRepo) ListCategories(ctx context.Context) ([]*models.Category, error) {
	args := m.Called(ctx)
	categories, _ := args.Get(0).([]*models.Category)
	return categories, args.Error(1)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepo) AddUser(ctx context.Context, user *models.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepo) GetUser(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *mockUserRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type mockImageRepo struct {
	mock.Mock
}

func (m *mockImageRepo) AddImage(ctx context.Context, img *models.Image) (int, error) {
	args := m.Called(ctx, img)
	return args.Int(0), args.Error(1)
}

func (m *mockImageRepo) GetImage(ctx context.Context, id int) (*models.Image, error) {
	args := m.Called(ctx, id)
	img, _ := args.Get(0).(*models.Image)
	return img, args.Error(1)
}

func (m *mockImageRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type mockUOW struct {
	mock.Mock
}

func (m *mockUOW) BeginTx(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	nextCtx, _ := args.Get(0).(context.Context)
	return nextCtx, args.Error(1)
}

func (m *mockUOW) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockUOW) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func testEventModel() *models.Event {
	lat, lon := 55.75, 37.62
	return &models.Event{
		ID:        0,
		CreatorID: 10,
		Name:      "Meetup title",
		EventTime: time.Now().UTC().Add(24 * time.Hour),
		City:      "Moscow",
		Location: models.Location{
			Name:      "Square",
			Latitude:  lat,
			Longitude: lon,
			Address:   "Red Square 1",
		},
	}
}

func testCreateEventReq() *dto.CreateEventRequest {
	lat, lon := 55.75, 37.62
	return &dto.CreateEventRequest{
		Name:      "Meetup title",
		EventTime: time.Now().UTC().Add(24 * time.Hour),
		City:      "Moscow",
		Location: &dto.EventLocation{
			Name:      "Square",
			Latitude:  &lat,
			Longitude: &lon,
			Address:   "Red Square 1",
		},
	}
}

func testUpdateEventReq() *dto.UpdateEventRequest {
	req := testCreateEventReq()
	return &dto.UpdateEventRequest{
		ID:              3,
		Name:            req.Name,
		Desc:            req.Desc,
		EventTime:       req.EventTime,
		TelegramChatURL: req.TelegramChatURL,
		City:            req.City,
		Location:        req.Location,
		ImageID:         req.ImageID,
	}
}

func intPtr(v int) *int {
	return &v
}

func expectExistingFutureEvent(er *mockEventRepo, ctx context.Context, id int) {
	existing := testEventModel()
	existing.ID = id
	er.On("GetEvent", ctx, id).Return(existing, nil)
}

func TestEventService_CreateEvent(t *testing.T) {
	t.Run("success without image", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		ev.ImageID = nil

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		uow.On("Commit", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(true, nil)
		er.On("CreateEvent", ctx, mock.MatchedBy(func(e *models.Event) bool {
			return e.CreatorID == 10 && e.ImageID == nil && !e.CreatedAt.IsZero() &&
				e.UpdatedAt.Equal(e.CreatedAt) && e.Deleted == false
		})).Return(100, nil)

		id, err := svc.CreateEvent(ctx, ev, 10)
		require.NoError(t, err)
		assert.Equal(t, 100, id)
		ur.AssertExpectations(t)
		er.AssertExpectations(t)
		ir.AssertNotCalled(t, "CheckExists", mock.Anything)
	})

	t.Run("success with image", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		ev.ImageID = intPtr(5)

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		uow.On("Commit", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(true, nil)
		ir.On("CheckExists", ctx, 5).Return(true, nil)
		er.On("CreateEvent", ctx, mock.MatchedBy(func(e *models.Event) bool {
			return e.ImageID != nil && *e.ImageID == 5
		})).Return(200, nil)

		id, err := svc.CreateEvent(ctx, ev, 10)
		require.NoError(t, err)
		assert.Equal(t, 200, id)
	})

	t.Run("creator not found", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(false, nil)

		id, err := svc.CreateEvent(ctx, ev, 10)
		assert.Zero(t, id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotExists))
		er.AssertNotCalled(t, "CreateEvent", mock.Anything, mock.Anything)
	})

	t.Run("user repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		repoErr := errors.New("db down")

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(false, repoErr)

		id, err := svc.CreateEvent(ctx, ev, 10)
		assert.Zero(t, id)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("image not found", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		ev.ImageID = intPtr(7)

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(true, nil)
		ir.On("CheckExists", ctx, 7).Return(false, nil)

		id, err := svc.CreateEvent(ctx, ev, 10)
		assert.Zero(t, id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotExists))
		er.AssertNotCalled(t, "CreateEvent", mock.Anything, mock.Anything)
	})

	t.Run("image repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		ev.ImageID = intPtr(7)
		repoErr := errors.New("timeout")

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(true, nil)
		ir.On("CheckExists", ctx, 7).Return(false, repoErr)

		id, err := svc.CreateEvent(ctx, ev, 10)
		assert.Zero(t, id)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("event repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir, uow := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo), new(mockUOW)
		svc := service.NewEventService(uow, er, ur, ir)
		ev := testCreateEventReq()
		ev.ImageID = nil
		repoErr := errors.New("insert failed")

		uow.On("BeginTx", ctx).Return(ctx, nil)
		uow.On("Rollback", ctx).Return(nil)
		ur.On("CheckExists", ctx, 10).Return(true, nil)
		er.On("CreateEvent", ctx, mock.Anything).Return(0, repoErr)

		id, err := svc.CreateEvent(ctx, ev, 10)
		assert.Zero(t, id)
		require.ErrorIs(t, err, repoErr)
	})
}

func TestEventService_GetEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		want := testEventModel()
		want.ID = 1

		er.On("GetEvent", ctx, 1).Return(want, nil)

		got, err := svc.GetEvent(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, want.CreatorID, got.CreatorID)
		assert.Equal(t, want.Name, got.Name)
	})

	t.Run("not found sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)

		er.On("GetEvent", ctx, 99).Return(nil, sql.ErrNoRows)

		got, err := svc.GetEvent(ctx, 99)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotFound))
	})

	t.Run("other repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		repoErr := errors.New("read error")

		er.On("GetEvent", ctx, 2).Return(nil, repoErr)

		got, err := svc.GetEvent(ctx, 2)
		assert.Nil(t, got)
		require.ErrorIs(t, err, repoErr)
	})
}

func TestEventService_UpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ImageID = nil

		expectExistingFutureEvent(er, ctx, ev.ID)
		er.On("UpdateEvent", ctx, mock.MatchedBy(func(e *models.Event) bool {
			return e.ID == 3 && !e.UpdatedAt.IsZero()
		})).Return(nil)

		require.NoError(t, svc.UpdateEvent(ctx, ev))
	})

	t.Run("success with image", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ImageID = intPtr(8)

		expectExistingFutureEvent(er, ctx, ev.ID)
		ir.On("CheckExists", ctx, 8).Return(true, nil)
		er.On("UpdateEvent", ctx, mock.Anything).Return(nil)

		require.NoError(t, svc.UpdateEvent(ctx, ev))
	})

	t.Run("image not found", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ImageID = intPtr(4)

		expectExistingFutureEvent(er, ctx, ev.ID)
		ir.On("CheckExists", ctx, 4).Return(false, nil)

		err := svc.UpdateEvent(ctx, ev)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotExists))
	})

	t.Run("image repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ImageID = intPtr(9)
		repoErr := errors.New("image lookup failed")

		expectExistingFutureEvent(er, ctx, ev.ID)
		ir.On("CheckExists", ctx, 9).Return(false, repoErr)

		err := svc.UpdateEvent(ctx, ev)
		require.ErrorIs(t, err, repoErr)
		er.AssertNotCalled(t, "UpdateEvent", mock.Anything, mock.Anything)
	})

	t.Run("event repo error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ImageID = nil
		repoErr := errors.New("update failed")

		expectExistingFutureEvent(er, ctx, ev.ID)
		er.On("UpdateEvent", ctx, mock.Anything).Return(repoErr)

		err := svc.UpdateEvent(ctx, ev)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("update no rows", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		ev := testUpdateEventReq()
		ev.ID = 5
		ev.ImageID = nil

		expectExistingFutureEvent(er, ctx, ev.ID)
		er.On("UpdateEvent", ctx, mock.Anything).Return(sql.ErrNoRows)

		err := svc.UpdateEvent(ctx, ev)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotFound))
	})
}

func TestEventService_DeleteEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)

		er.On("DeleteEvent", ctx, 1).Return(nil)

		require.NoError(t, svc.DeleteEvent(ctx, 1))
	})

	t.Run("not found from repo", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)

		er.On("DeleteEvent", ctx, 2).Return(exceptions.ErrNotFound)

		err := svc.DeleteEvent(ctx, 2)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotFound))
	})

	t.Run("already deleted from repo", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)

		er.On("DeleteEvent", ctx, 3).Return(exceptions.ErrAlreadyDeleted)

		err := svc.DeleteEvent(ctx, 3)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrAlreadyDeleted))
	})

	t.Run("other error", func(t *testing.T) {
		ctx := context.Background()
		er, ur, ir := new(mockEventRepo), new(mockUserRepo), new(mockImageRepo)
		svc := service.NewEventService(new(mockUOW), er, ur, ir)
		repoErr := errors.New("constraint")

		er.On("DeleteEvent", ctx, 4).Return(repoErr)

		err := svc.DeleteEvent(ctx, 4)
		require.ErrorIs(t, err, repoErr)
	})
}
