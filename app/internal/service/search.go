package service

import (
	"JoinUp/internal/models"
	"context"
)

type SearchEngine struct {
	eventRepo EventRepo
}

func NewSearchEngine(eventRepo EventRepo) SearchEngine {
	return SearchEngine{eventRepo: eventRepo}
}

func (s *SearchEngine) SearchEvents(ctx context.Context, filter models.EventSearchFilter) ([]*models.Event, error) {
	return s.eventRepo.SearchEvents(ctx, filter)
}
