//go:build allure

package testutil

import (
	"JoinUp/internal/models"
	"time"
)

type ReviewBuilder struct {
	review models.Review
}

func NewReviewBuilder() *ReviewBuilder {
	return &ReviewBuilder{review: models.Review{
		EventID:   7,
		AuthorID:  20,
		Rating:    5,
		Text:      "Отличная встреча",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		Author:    models.ReviewAuthor{ID: 20, Name: "Алексей"},
	}}
}

func (b *ReviewBuilder) WithRating(rating int) *ReviewBuilder {
	b.review.Rating = rating
	return b
}

func (b *ReviewBuilder) WithText(text string) *ReviewBuilder {
	b.review.Text = text
	return b
}

func (b *ReviewBuilder) Build() *models.Review {
	review := b.review
	return &review
}
