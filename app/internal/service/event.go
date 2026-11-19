package service

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type EventService struct {
	eventRepo EventRepo
	userRepo  UserRepo
	imageRepo ImageRepo
	uow       UOW
}

func NewEventService(uow UOW, eventRepo EventRepo, userRepo UserRepo, imageRepo ImageRepo) EventService {
	return EventService{uow: uow, eventRepo: eventRepo, userRepo: userRepo, imageRepo: imageRepo}
}

func (s *EventService) CreateEvent(ctx context.Context, event models.Event, userID int) (int, error) {
	if eventAlreadyPassed(event.EventTime) {
		return 0, fmt.Errorf("%w: event time must not be in the past", exceptions.ErrEventAlreadyPassed)
	}

	event.CreatedAt = time.Now().UTC()
	event.UpdatedAt = event.CreatedAt
	event.Deleted = false
	event.CreatorID = userID
	event.Members = []models.EventMember{}
	event.Categories = []models.Category{}

	ctx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer s.uow.Rollback(ctx)

	ok, err := s.userRepo.CheckExists(ctx, event.CreatorID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("%w: creator with such id not found, creator_id=%d", exceptions.ErrNotExists, event.CreatorID)
	}

	if event.ImageID != nil {
		ok, err := s.imageRepo.CheckExists(ctx, *event.ImageID)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, fmt.Errorf("%w: image with such id not found, image_id=%d", exceptions.ErrNotExists, *event.ImageID)
		}
	}

	id, err := s.eventRepo.CreateEvent(ctx, &event)
	if err != nil {
		return 0, err
	}

	return id, s.uow.Commit(ctx)
}

func (s *EventService) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	event, err := s.eventRepo.GetEvent(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, id)
	} else if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, event models.Event, categoryIDs []int) error {
	now := time.Now().UTC()
	ctx, err := s.uow.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer s.uow.Rollback(ctx)

	oldEvent, err := s.eventRepo.GetEvent(ctx, event.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, event.ID)
	}
	if err != nil {
		return err
	}
	if oldEvent.EventTime.Before(now) || event.EventTime.Before(now) {
		return fmt.Errorf("%w: event time must not be in the past", exceptions.ErrEventAlreadyPassed)
	}

	if event.ImageID != nil {
		ok, err := s.imageRepo.CheckExists(ctx, *event.ImageID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: image with such id not found, image_id=%d", exceptions.ErrNotExists, *event.ImageID)
		}
	}

	if categoryIDs != nil {
		categories, err := s.eventRepo.ListCategories(ctx)
		if err != nil {
			return err
		}
		byID := make(map[int]*models.Category, len(categories))
		for _, category := range categories {
			byID[category.ID] = category
		}
		event.Categories = make([]models.Category, 0, len(categoryIDs))
		for _, id := range categoryIDs {
			category, ok := byID[id]
			if !ok {
				return fmt.Errorf("%w: category with id = %d not found", exceptions.ErrNotExists, id)
			}
			event.Categories = append(event.Categories, *category)
		}
	}

	event.UpdatedAt = now

	if err := s.eventRepo.UpdateEvent(ctx, &event); errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, event.ID)
	} else if err != nil {
		return err
	}

	return s.uow.Commit(ctx)
}

func (s *EventService) DeleteEvent(ctx context.Context, id int) error {
	err := s.eventRepo.DeleteEvent(ctx, id)
	if errors.Is(err, exceptions.ErrNotFound) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, id)
	} else if errors.Is(err, exceptions.ErrAlreadyDeleted) {
		return fmt.Errorf("%w: event with id = %d is already deleted", exceptions.ErrAlreadyDeleted, id)
	}
	return err
}

func (s *EventService) JoinEvent(ctx context.Context, eventID int, userID int) error {
	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return err
	}
	if eventAlreadyPassed(event.EventTime) {
		return fmt.Errorf("%w: event id = %d already passed", exceptions.ErrEventAlreadyPassed, eventID)
	}

	ok, err := s.userRepo.CheckExists(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: user with id = %d not found", exceptions.ErrNotExists, userID)
	}

	err = s.eventRepo.JoinEvent(ctx, eventID, userID)
	if errors.Is(err, exceptions.ErrNotFound) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		return fmt.Errorf("%w: user id = %d already participates in event id = %d", exceptions.ErrAlreadyExists, userID, eventID)
	}
	return err
}

func (s *EventService) LeaveEvent(ctx context.Context, eventID int, userID int) error {
	err := s.eventRepo.LeaveEvent(ctx, eventID, userID)
	if errors.Is(err, exceptions.ErrNotFound) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if errors.Is(err, exceptions.ErrNotExists) {
		return fmt.Errorf("%w: user id = %d does not participate in event id = %d", exceptions.ErrNotExists, userID, eventID)
	}
	return err
}

func (s *EventService) ListJoinedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	return s.eventRepo.ListJoinedEvents(ctx, userID)
}

func (s *EventService) ListCreatedEvents(ctx context.Context, userID int) ([]*models.Event, error) {
	return s.eventRepo.ListCreatedEvents(ctx, userID)
}

func (s *EventService) UploadEventImage(ctx context.Context, eventID int, name string, data []byte) (int, error) {
	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return 0, err
	}
	if eventAlreadyPassed(event.EventTime) {
		return 0, fmt.Errorf("%w: event id = %d already passed", exceptions.ErrEventAlreadyPassed, eventID)
	}

	ctx, err = s.uow.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer s.uow.Rollback(ctx)

	imageID, err := s.imageRepo.AddImage(ctx, &models.Image{Name: name, Data: data})
	if err != nil {
		return 0, err
	}

	err = s.eventRepo.UpdateEventImage(ctx, eventID, imageID, time.Now().UTC())
	if errors.Is(err, exceptions.ErrNotFound) {
		return 0, fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return 0, err
	}

	return imageID, s.uow.Commit(ctx)
}

func (s *EventService) AddEventCategory(ctx context.Context, eventID, categoryID int) error {
	event, err := s.eventRepo.GetEvent(ctx, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if err != nil {
		return err
	}
	if eventAlreadyPassed(event.EventTime) {
		return fmt.Errorf("%w: event id = %d already passed", exceptions.ErrEventAlreadyPassed, eventID)
	}

	err = s.eventRepo.AddEventCategory(ctx, eventID, categoryID)
	if errors.Is(err, exceptions.ErrNotFound) {
		return fmt.Errorf("%w: no event with id = %d", exceptions.ErrNotFound, eventID)
	}
	if errors.Is(err, exceptions.ErrNotExists) {
		return fmt.Errorf("%w: category with id = %d not found", exceptions.ErrNotExists, categoryID)
	}
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		return fmt.Errorf("%w: category id = %d already added to event id = %d", exceptions.ErrAlreadyExists, categoryID, eventID)
	}
	return err
}

func (s *EventService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return s.eventRepo.ListCategories(ctx)
}

func eventAlreadyPassed(eventTime time.Time) bool {
	return eventTime.UTC().Before(time.Now().UTC())
}
