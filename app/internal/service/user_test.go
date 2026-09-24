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
	appjwt "JoinUp/internal/utils/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func testCreateUserReq() *dto.CreateUserRequest {
	return &dto.CreateUserRequest{
		Name:     "Misha",
		Age:      22,
		Login:    "admin",
		Password: "admin123",
		City:     "Moscow",
	}
}

func testUserModel() *models.User {
	return &models.User{
		ID:        1,
		Name:      "Misha",
		Age:       22,
		Login:     "admin",
		Password:  "$2a$10$abcdefghijklmnopqrstuv",
		CreatedAt: time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
		City:      "Moscow",
		Role:      "user",
	}
}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)
		req := testCreateUserReq()

		ur.On("AddUser", ctx, mock.MatchedBy(func(user *models.User) bool {
			if user.Login != "admin" || user.City != "Moscow" || user.CreatedAt.IsZero() {
				return false
			}
			return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("admin123")) == nil
		})).Return(10, nil)

		id, err := svc.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 10, id)
	})

	t.Run("duplicate login", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)
		req := testCreateUserReq()

		ur.On("AddUser", ctx, mock.Anything).Return(0, exceptions.ErrAlreadyExists)

		id, err := svc.CreateUser(ctx, req)
		assert.Zero(t, id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrAlreadyExists))
	})
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)
		user := testUserModel()

		ur.On("GetUser", ctx, 1).Return(user, nil)

		got, err := svc.GetUser(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, got.ID)
		assert.Equal(t, "admin", got.Login)
		assert.Equal(t, "user", got.Role)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)

		ur.On("GetUser", ctx, 99).Return(nil, sql.ErrNoRows)

		got, err := svc.GetUser(ctx, 99)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrNotFound))
	})
}

func TestUserService_Auth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)

		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		ur.On("GetUserByLogin", ctx, "admin").Return(&models.User{
			ID:       1,
			Login:    "admin",
			Password: string(hash),
			Role:     "user",
		}, nil)

		token, err := svc.Auth(ctx, "admin", "admin123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		ctx := context.Background()
		ur, ir := new(mockUserRepo), new(mockImageRepo)
		jwtMgr := appjwt.NewJwtManager("test-secret")
		svc := service.NewUserService(ur, ir, nil, &jwtMgr)

		ur.On("GetUserByLogin", ctx, "admin").Return(nil, sql.ErrNoRows)

		token, err := svc.Auth(ctx, "admin", "wrong")
		assert.Empty(t, token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, exceptions.ErrInvalidCredentials))
	})
}
