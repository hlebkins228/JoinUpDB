package dto

import "time"

type CreateReviewRequest struct {
	EventID int    `param:"id" validate:"required,gte=1"`
	Rating  int    `json:"rating" validate:"required,gte=1,lte=5"`
	Text    string `json:"text" validate:"required,max=2000"`
}

type ReviewIDRequest struct {
	EventID int `param:"id" validate:"required,gte=1"`
}

type ReviewAuthor struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	AvatarID *int   `json:"avatar_id"`
}

type ReviewResponse struct {
	ID        int           `json:"id"`
	EventID   int           `json:"event_id"`
	AuthorID  int           `json:"author_id"`
	Rating    int           `json:"rating"`
	Text      string        `json:"text"`
	CreatedAt time.Time     `json:"created_at"`
	Author    *ReviewAuthor `json:"author"`
}

type ReviewsResponse struct {
	Reviews []*ReviewResponse `json:"reviews"`
}
