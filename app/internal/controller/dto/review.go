package dto

import (
	"JoinUp/internal/models"
	"time"
)

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

func ReviewsFromModels(reviews []*models.Review) *ReviewsResponse {
	response := &ReviewsResponse{Reviews: make([]*ReviewResponse, 0, len(reviews))}
	for _, review := range reviews {
		response.Reviews = append(response.Reviews, ReviewFromModel(review))
	}
	return response
}

func ReviewFromModel(r *models.Review) *ReviewResponse {
	return &ReviewResponse{
		ID:        r.ID,
		EventID:   r.EventID,
		AuthorID:  r.AuthorID,
		Rating:    r.Rating,
		Text:      r.Text,
		CreatedAt: r.CreatedAt,
		Author: &ReviewAuthor{
			ID:       r.Author.ID,
			Name:     r.Author.Name,
			AvatarID: r.Author.AvatarID,
		},
	}
}
