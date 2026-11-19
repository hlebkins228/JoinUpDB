package service

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	appjwt "JoinUp/internal/utils/jwt"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo  UserRepo
	imageRepo ImageRepo
	eventRepo EventRepo
	jwtMgr    *appjwt.JwtManager
}

func NewUserService(userRepo UserRepo, imageRepo ImageRepo, eventRepo EventRepo, jwtMgr *appjwt.JwtManager) UserService {
	return UserService{userRepo: userRepo, imageRepo: imageRepo, eventRepo: eventRepo, jwtMgr: jwtMgr}
}

func (s *UserService) CreateUser(ctx context.Context, user models.User) (int, error) {
	user.CreatedAt = time.Now().UTC()
	user.Role = "user"

	if user.AvatarID != nil {
		ok, err := s.imageRepo.CheckExists(ctx, *user.AvatarID)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, fmt.Errorf("%w: image with such id not found, image_id=%d", exceptions.ErrNotExists, *user.AvatarID)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = string(hash)

	id, err := s.userRepo.AddUser(ctx, &user)
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		return 0, fmt.Errorf("%w: user with login `%s` already exists", exceptions.ErrAlreadyExists, user.Login)
	}
	return id, err
}

func (s *UserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userRepo.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: no user with id = %d", exceptions.ErrNotFound, id)
	}
	if err != nil {
		return nil, err
	}

	if user.OrganizerRating != nil {
		rating := roundHalfUpTenth(*user.OrganizerRating)
		user.OrganizerRating = &rating
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, user models.User) error {
	user.ID = id

	if user.AvatarID != nil {
		ok, err := s.imageRepo.CheckExists(ctx, *user.AvatarID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: image with such id not found, image_id=%d", exceptions.ErrNotExists, *user.AvatarID)
		}
	}

	err := s.userRepo.UpdateUser(ctx, &user)
	if errors.Is(err, exceptions.ErrNotFound) {
		return fmt.Errorf("%w: no user with id = %d", exceptions.ErrNotFound, id)
	}
	return err
}

func (s *UserService) UploadImage(ctx context.Context, name string, data []byte) (int, error) {
	return s.imageRepo.AddImage(ctx, &models.Image{
		Name: name,
		Data: data,
	})
}

func (s *UserService) GetImage(ctx context.Context, id int) (*models.Image, error) {
	image, err := s.imageRepo.GetImage(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: no image with id = %d", exceptions.ErrNotFound, id)
	}
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *UserService) Auth(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: invalid login or password", exceptions.ErrInvalidCredentials)
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("%w: invalid login or password", exceptions.ErrInvalidCredentials)
	}

	return s.jwtMgr.NewToken(user.ID, user.Role)
}

func (s *UserService) CheckPerms(ctx context.Context, eventID int, userID int, role string) (bool, error) {
	if role == "admin" {
		return true, nil
	}

	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return false, err
	}

	return event.CreatorID == userID, nil
}

// roundHalfUpTenth округляет вещественное число до одного знака после запятой
// по правилу математического округления half-up (0.05 → 0.1).
func roundHalfUpTenth(x float64) float64 {
	return math.Floor(x*10+0.5) / 10
}
