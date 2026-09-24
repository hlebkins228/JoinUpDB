package dto

import (
	"JoinUp/internal/models"
	"strconv"
	"strings"
	"time"
)

type CreateEventRequest struct {
	Name            string         `json:"name" validate:"required,min=5,max=100"`
	Desc            string         `json:"desc,omitempty"`
	EventTime       time.Time      `json:"event_time" validate:"required"`
	TelegramChatURL string         `json:"telegram_chat_url,omitempty" validate:"omitempty,url"`
	City            string         `json:"city" validate:"required"`
	Location        *EventLocation `json:"location" validate:"required"`
	ImageID         *int           `json:"image_id,omitempty"`
}

type UpdateEventRequest struct {
	ID              int            `param:"id" validate:"required,gte=1"`
	Name            string         `json:"name" validate:"required,min=5,max=100"`
	Desc            string         `json:"desc,omitempty"`
	EventTime       time.Time      `json:"event_time" validate:"required"`
	TelegramChatURL string         `json:"telegram_chat_url,omitempty" validate:"omitempty,url"`
	City            string         `json:"city" validate:"required"`
	Location        *EventLocation `json:"location" validate:"required"`
	ImageID         *int           `json:"image_id,omitempty"`
}

type EventSearchRequest struct {
	Name        *string    `query:"name"`
	EventFrom   *time.Time `query:"event_from"`
	EventTo     *time.Time `query:"event_to"`
	City        *string    `query:"city"`
	CategoryIDs []int      `query:"category_id"`
}

type AddEventCategoryRequest struct {
	EventID    int `param:"id" validate:"required,gte=1"`
	CategoryID int `json:"category_id" validate:"required,gte=1"`
}

type EventLocation struct {
	Name      string   `json:"name" validate:"required"`
	Longitude *float64 `json:"longitude" validate:"required"`
	Latitude  *float64 `json:"latitude" validate:"required"`
	Address   string   `json:"address" validate:"required"`
}

type EventIDRequest struct {
	ID int `param:"id" validate:"required,gte=1"`
}

type EventResponse struct {
	ID              int                 `json:"id"`
	CreatorID       int                 `json:"creator_id"`
	Name            string              `json:"name"`
	Desc            string              `json:"desc,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	EventTime       time.Time           `json:"event_time"`
	TelegramChatURL string              `json:"telegram_chat_url,omitempty"`
	City            string              `json:"city"`
	Creator         *EventMember        `json:"creator,omitempty"`
	Members         []*EventMember      `json:"members,omitempty"`
	Categories      []*CategoryResponse `json:"categories,omitempty"`
	Location        *EventLocation      `json:"location"`
	ImageID         *int                `json:"image_id,omitempty"`
	Deleted         bool                `json:"deleted"`
}

type EventIDResponse struct {
	ID int `json:"id"`
}

type EventsResponse struct {
	Events []*EventResponse `json:"events"`
}

type ImageIDResponse struct {
	ID int `json:"id"`
}

type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EventMember struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	AvatarID *int   `json:"avatar_id,omitempty"`
}

type CategoriesResponse struct {
	Categories []*CategoryResponse `json:"categories"`
}

func (r *CreateEventRequest) ToModel() models.Event {
	return models.Event{
		Name:            r.Name,
		Desc:            r.Desc,
		EventTime:       r.EventTime.UTC(),
		TelegramChatURL: r.TelegramChatURL,
		City:            r.City,
		Location:        r.Location.ToModel(),
		ImageID:         r.ImageID,
	}
}

func (r *UpdateEventRequest) ToModel() models.Event {
	return models.Event{
		ID:              r.ID,
		Name:            r.Name,
		Desc:            r.Desc,
		EventTime:       r.EventTime.UTC(),
		TelegramChatURL: r.TelegramChatURL,
		City:            r.City,
		Location:        r.Location.ToModel(),
		ImageID:         r.ImageID,
	}
}

func (r *EventSearchRequest) ToModel() models.EventSearchFilter {
	return models.EventSearchFilter{
		Name:        emptyStringToNil(r.Name),
		EventFrom:   timeToUTC(r.EventFrom),
		EventTo:     timeToUTC(r.EventTo),
		City:        emptyStringToNil(r.City),
		CategoryIDs: r.CategoryIDs,
	}
}

func ParseCategoryIDs(values []string) ([]int, error) {
	ids := make([]int, 0)
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (l *EventLocation) ToModel() models.Location {
	return models.Location{
		Name:      l.Name,
		Longitude: *l.Longitude,
		Latitude:  *l.Latitude,
		Address:   l.Address,
	}
}

func emptyStringToNil(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func timeToUTC(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
