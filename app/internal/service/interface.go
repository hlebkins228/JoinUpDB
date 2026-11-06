package service

import (
	"JoinUp/internal/models"
	"context"
	"time"
)

type EventRepo interface {
	CreateEvent(ctx context.Context, event *models.Event) (int, error)
	GetEvent(ctx context.Context, id int) (*models.Event, error)
	UpdateEvent(ctx context.Context, event *models.Event) error
	DeleteEvent(ctx context.Context, id int) error
	JoinEvent(ctx context.Context, eventID int, userID int) error
	LeaveEvent(ctx context.Context, eventID int, userID int) error
	IsParticipant(ctx context.Context, eventID, userID int) (bool, error)
	UpdateEventImage(ctx context.Context, eventID int, imageID int, updatedAt time.Time) error
	AddEventCategory(ctx context.Context, eventID int, categoryID int) error
	SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error)
	ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error)
	ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type ReviewRepo interface {
	Create(ctx context.Context, rv *models.Review) (*models.Review, error)
	ListByEvent(ctx context.Context, eventID int) ([]*models.Review, error)
	ExistsByEventAndAuthor(ctx context.Context, eventID, authorID int) (bool, error)
}

type ImageRepo interface {
	AddImage(ctx context.Context, img *models.Image) (int, error)
	GetImage(ctx context.Context, id int) (*models.Image, error)
	CheckExists(ctx context.Context, id int) (bool, error)
}

type UserRepo interface {
	CheckExists(ctx context.Context, id int) (bool, error)
	AddUser(ctx context.Context, user *models.User) (int, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type UOW interface {
	BeginTx(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
