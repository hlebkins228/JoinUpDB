package handlers

import (
	"JoinUp/internal/models"
	"JoinUp/internal/service"
	"context"
)

type EventService interface {
	CreateEvent(ctx context.Context, event models.Event, userID int) (int, error)
	GetEvent(ctx context.Context, id int) (*models.Event, error)
	UpdateEvent(ctx context.Context, event models.Event, categoryIDs []int) error
	DeleteEvent(ctx context.Context, id int) error
	JoinEvent(ctx context.Context, eventID int, userID int) error
	LeaveEvent(ctx context.Context, eventID int, userID int) error
	ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error)
	ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error)
	UploadEventImage(ctx context.Context, eventID int, name string, data []byte) (int, error)
	AddEventCategory(ctx context.Context, eventID, categoryID int) error
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type UserService interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, id int, user models.User) error
	UploadImage(ctx context.Context, name string, data []byte) (int, error)
	GetImage(ctx context.Context, id int) (*models.Image, error)
	Auth(ctx context.Context, login, password string) (string, error)
	CheckPerms(ctx context.Context, eventID int, userID int, role string) (bool, error)
}

type SearchEngine interface {
	SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error)
}

type ReviewService interface {
	CreateReview(ctx context.Context, eventID, authorID, rating int, text string) (*models.Review, error)
	ListReviews(ctx context.Context, eventID int) (*service.ReviewList, error)
}
