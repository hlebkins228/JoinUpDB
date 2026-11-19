package postgres

import (
	"JoinUp/internal/models"
	"time"
)

type locationRow struct {
	name      string
	longitude float64
	latitude  float64
	address   string
}

type eventRow struct {
	id              int
	creatorID       int
	name            string
	description     string
	createdAt       time.Time
	updatedAt       time.Time
	eventDate       time.Time
	telegramChatURL string
	city            string
	location        locationRow
	imageID         *int
	deleted         bool
}

func (r eventRow) toModel() *models.Event {
	return &models.Event{
		ID:              r.id,
		CreatorID:       r.creatorID,
		Name:            r.name,
		Desc:            r.description,
		CreatedAt:       r.createdAt,
		UpdatedAt:       r.updatedAt,
		EventTime:       r.eventDate,
		TelegramChatURL: r.telegramChatURL,
		City:            r.city,
		Location: models.Location{
			Name:      r.location.name,
			Longitude: r.location.longitude,
			Latitude:  r.location.latitude,
			Address:   r.location.address,
		},
		ImageID: r.imageID,
		Deleted: r.deleted,
	}
}

type userRow struct {
	id            int
	name          string
	age           int
	login         string
	passwordHash  string
	createdAt     time.Time
	city          string
	telegramLogin *string
	avatarID      *int
	role          string
}

func (r userRow) toModel() *models.User {
	return &models.User{
		ID:            r.id,
		Name:          r.name,
		Age:           r.age,
		Login:         r.login,
		Password:      r.passwordHash,
		CreatedAt:     r.createdAt,
		City:          r.city,
		TelegramLogin: r.telegramLogin,
		AvatarID:      r.avatarID,
		Role:          r.role,
	}
}

type reviewRow struct {
	id        int
	eventID   int
	authorID  int
	rating    int
	text      string
	createdAt time.Time
}

type userSummaryRow struct {
	id       int
	name     string
	avatarID *int
}

type categoryRow struct {
	id   int
	name string
}

type imageRow struct {
	name string
	data []byte
}
