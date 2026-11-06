package models

import "time"

type Review struct {
	ID        int
	EventID   int
	AuthorID  int
	Rating    int
	Text      string
	CreatedAt time.Time
	Author    ReviewAuthor
}

type ReviewAuthor struct {
	ID       int
	Name     string
	AvatarID *int
}
