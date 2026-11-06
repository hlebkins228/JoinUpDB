package models

import "time"

type Event struct {
	ID              int
	CreatorID       int
	Name            string
	Desc            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EventTime       time.Time
	TelegramChatURL string
	City            string
	Creator         EventMember
	Members         []EventMember
	Categories      []Category
	Location        Location
	ImageID         *int
	Deleted         bool
}

type EventSearchFilter struct {
	Name        *string
	EventFrom   *time.Time
	EventTo     *time.Time
	City        *string
	CategoryIDs []int
}

type EventMember struct {
	ID       int
	Name     string
	AvatarID *int
}
