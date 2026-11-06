package service

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	reviewRatingMin = 1
	reviewRatingMax = 5
	reviewTextMin   = 1
	reviewTextMax   = 2000
)

type ReviewService struct {
	reviewRepo ReviewRepo
	eventRepo  EventRepo
	uow        UOW
}

func NewReviewService(uow UOW, reviewRepo ReviewRepo, eventRepo EventRepo) ReviewService {
	return ReviewService{uow: uow, reviewRepo: reviewRepo, eventRepo: eventRepo}
}

func (s *ReviewService) CreateReview(
	ctx context.Context, eventID, authorID, rating int, text string,
) (*dto.ReviewResponse, error) {
	trimmedText := strings.TrimSpace(text)
	if rating < reviewRatingMin || rating > reviewRatingMax {
		return nil, fmt.Errorf("%w: rating must be in [%d..%d], got %d",
			exceptions.ErrInvalidPayload, reviewRatingMin, reviewRatingMax, rating)
	}
	textLen := utf8.RuneCountInString(trimmedText)
	if textLen < reviewTextMin || textLen > reviewTextMax {
		return nil, fmt.Errorf("%w: text length must be in [%d..%d] runes, got %d",
			exceptions.ErrInvalidPayload, reviewTextMin, reviewTextMax, textLen)
	}

	ctx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.uow.Rollback(ctx)

	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return nil, err
	}
	if event.Deleted {
		return nil, fmt.Errorf("%w: event with id = %d is deleted", exceptions.ErrNotFound, eventID)
	}

	if event.EventTime.UTC().After(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: event id = %d is in the future", exceptions.ErrEventNotPassed, eventID)
	}

	if event.CreatorID == authorID {
		return nil, fmt.Errorf("%w: user id = %d is organizer of event id = %d",
			exceptions.ErrOrganizerCannotReview, authorID, eventID)
	}

	exists, err := s.reviewRepo.ExistsByEventAndAuthor(ctx, eventID, authorID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: review by user id = %d for event id = %d already exists",
			exceptions.ErrAlreadyExists, authorID, eventID)
	}

	review := &models.Review{
		EventID:   eventID,
		AuthorID:  authorID,
		Rating:    rating,
		Text:      trimmedText,
		CreatedAt: time.Now().UTC(),
	}

	created, err := s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}

	if err := s.uow.Commit(ctx); err != nil {
		return nil, err
	}

	return modelToReviewResponseDTO(created), nil
}

func (s *ReviewService) ListReviews(ctx context.Context, eventID int) (*dto.ReviewsResponse, error) {
	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return nil, err
	}
	if event.Deleted {
		return nil, fmt.Errorf("%w: event with id = %d is deleted", exceptions.ErrNotFound, eventID)
	}

	reviews, err := s.reviewRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	resp := &dto.ReviewsResponse{Reviews: make([]*dto.ReviewResponse, 0, len(reviews))}
	for _, r := range reviews {
		resp.Reviews = append(resp.Reviews, modelToReviewResponseDTO(r))
	}

	return resp, nil
}

func modelToReviewResponseDTO(r *models.Review) *dto.ReviewResponse {
	return &dto.ReviewResponse{
		ID:        r.ID,
		EventID:   r.EventID,
		AuthorID:  r.AuthorID,
		Rating:    r.Rating,
		Text:      r.Text,
		CreatedAt: r.CreatedAt,
		Author: &dto.ReviewAuthor{
			ID:       r.Author.ID,
			Name:     r.Author.Name,
			AvatarID: r.Author.AvatarID,
		},
	}
}
