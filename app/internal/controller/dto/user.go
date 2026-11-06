package dto

import (
	"JoinUp/internal/models"
	"time"
)

type CreateUserRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	Age           int     `json:"age" validate:"required,gte=1"`
	Login         string  `json:"login" validate:"required,min=3,max=100"`
	Password      string  `json:"password" validate:"required,min=6,max=255"`
	City          string  `json:"city" validate:"required"`
	TelegramLogin *string `json:"telegram_login,omitempty"`
	AvatarID      *int    `json:"avatar_id,omitempty"`
}

type UserIDRequest struct {
	ID int `param:"id" validate:"required,gte=1"`
}

type UpdateUserRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	Age           int     `json:"age" validate:"required,gte=1"`
	City          string  `json:"city" validate:"required"`
	TelegramLogin *string `json:"telegram_login,omitempty"`
	AvatarID      *int    `json:"avatar_id,omitempty"`
}

type UserIDResponse struct {
	ID int `json:"id"`
}

type UserResponse struct {
	ID                    int       `json:"id"`
	Name                  string    `json:"name"`
	Age                   int       `json:"age"`
	Login                 string    `json:"login"`
	CreatedAt             time.Time `json:"created_at"`
	City                  string    `json:"city"`
	TelegramLogin         *string   `json:"telegram_login,omitempty"`
	AvatarID              *int      `json:"avatar_id,omitempty"`
	Role                  string    `json:"role"`
	CreatedEventsCount    int       `json:"created_events_count"`
	JoinedEventsCount     int       `json:"joined_events_count"`
	OrganizerRating       *float64  `json:"organizer_rating"`
	OrganizerReviewsCount int       `json:"organizer_reviews_count"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (r *CreateUserRequest) ToModel() models.User {
	return models.User{
		Name:          r.Name,
		Age:           r.Age,
		Login:         r.Login,
		Password:      r.Password,
		City:          r.City,
		TelegramLogin: r.TelegramLogin,
		AvatarID:      r.AvatarID,
	}
}

func (r *UpdateUserRequest) ToModel() models.User {
	return models.User{
		Name:          r.Name,
		Age:           r.Age,
		City:          r.City,
		TelegramLogin: r.TelegramLogin,
		AvatarID:      r.AvatarID,
	}
}
