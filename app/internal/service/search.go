package service

import (
	"JoinUp/internal/controller/dto"
	"context"
)

type SearchEngine struct {
	eventRepo EventRepo
}

func NewSearchEngine(eventRepo EventRepo) SearchEngine {
	return SearchEngine{eventRepo: eventRepo}
}

func (s *SearchEngine) SearchEvents(ctx context.Context, req *dto.EventSearchRequest) (*dto.EventsResponse, error) {
	events, err := s.eventRepo.SearchEvents(ctx, req.ToModel())
	if err != nil {
		return nil, err
	}

	resp := &dto.EventsResponse{Events: make([]*dto.EventResponse, 0, len(events))}
	for _, event := range events {
		resp.Events = append(resp.Events, modelToEventResponseDTO(event))
	}

	return resp, nil
}
