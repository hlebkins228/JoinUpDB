package exceptions

import "errors"

var (
	ErrNotFound           = errors.New("not found error")
	ErrNotExists          = errors.New("not exists error")
	ErrAlreadyDeleted     = errors.New("already deleted error")
	ErrAlreadyExists      = errors.New("already exists error")
	ErrEventAlreadyPassed = errors.New("event already passed error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoTxObj            = errors.New("no transaction object in context")
	ErrTxType             = errors.New("unexpected tx type")
	ErrNoUserID           = errors.New("no user id in request context")

	ErrEventNotPassed        = errors.New("event not passed yet")
	ErrNotParticipant        = errors.New("user is not a participant of this event")
	ErrOrganizerCannotReview = errors.New("organizer cannot review own event")
	ErrInvalidPayload        = errors.New("invalid payload")
)
