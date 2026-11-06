package handlers

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"JoinUp/internal/utils"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type EventHandler struct {
	ctx      context.Context
	logger   *zap.Logger
	eventSvc EventService
	userSvc  UserService
	search   SearchEngine
}

func NewEventHandler(ctx context.Context, logger *zap.Logger, eventSvc EventService, userSvc UserService, search SearchEngine) EventHandler {
	return EventHandler{ctx: ctx, logger: logger, eventSvc: eventSvc, userSvc: userSvc, search: search}
}

// CreateEvent godoc
// @Summary Create event
// @Description Creates new event with location
// @Tags events
// @Accept json
// @Produce json
// @Param event body dto.CreateEventRequest true "Event payload"
// @Success 200 {object} dto.EventIDResponse
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event [post]
func (h *EventHandler) CreateEvent(c *echo.Context) error {
	userID, _, err := h.UserFromContext(c)
	if userID == nil || err != nil {
		return err
	}

	var req dto.CreateEventRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on create event: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}

	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on create event: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	id, err := h.eventSvc.CreateEvent(c.Request().Context(), &req, *userID)
	if errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on create event: reference not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	} else if errors.Is(err, exceptions.ErrEventAlreadyPassed) {
		h.logger.Info("error on create event: event time is in the past", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: err.Error()})
	} else if err != nil {
		h.logger.Error("error on create event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.EventIDResponse{ID: id})
}

// GetEvent godoc
// @Summary Get event by id
// @Description Returns event by identifier
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.EventResponse
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/user/event/{id} [get]
func (h *EventHandler) GetEvent(c *echo.Context) error {
	var req dto.EventIDRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on get event: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on get event: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	event, err := h.eventSvc.GetEvent(c.Request().Context(), req.ID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on get event: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	} else if err != nil {
		h.logger.Error("error on get event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, event)
}

// SearchEvents godoc
// @Summary Search events
// @Description Searches events by optional name substring, event datetime interval, city and categories
// @Tags events
// @Produce json
// @Param name query string false "Name substring"
// @Param event_from query string false "Event datetime from"
// @Param event_to query string false "Event datetime to"
// @Param city query string false "City"
// @Param category_id query []int false "Category ids"
// @Success 200 {object} dto.EventsResponse
// @Failure 400 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/user/event/search [get]
func (h *EventHandler) SearchEvents(c *echo.Context) error {
	var req dto.EventSearchRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on search events: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}

	categoryIDs, err := dto.ParseCategoryIDs(c.QueryParams()["category_id"])
	if err != nil {
		h.logger.Info("error on search events: invalid category_id", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "invalid category_id"})
	}
	if len(categoryIDs) == 0 {
		categoryIDs, err = dto.ParseCategoryIDs(c.QueryParams()["category_ids"])
		if err != nil {
			h.logger.Info("error on search events: invalid category_ids", zap.Error(err))
			return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "invalid category_ids"})
		}
	}
	req.CategoryIDs = uniquePositiveIDs(categoryIDs)

	resp, err := h.search.SearchEvents(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("error on search events", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, resp)
}

// ListCreatedEvents godoc
// @Summary List created events
// @Description Returns events created by the current user, including past ones
// @Tags events
// @Produce json
// @Success 200 {object} dto.EventsResponse
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/created [get]
func (h *EventHandler) ListCreatedEvents(c *echo.Context) error {
	userID, _, err := h.UserFromContext(c)
	if userID == nil || err != nil {
		return err
	}

	resp, err := h.eventSvc.ListCreatedEvents(c.Request().Context(), *userID)
	if err != nil {
		h.logger.Error("error on list created events", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, resp)
}

// ListJoinedEvents godoc
// @Summary List joined events
// @Description Returns events the current user participates in
// @Tags events
// @Produce json
// @Success 200 {object} dto.EventsResponse
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/joined [get]
func (h *EventHandler) ListJoinedEvents(c *echo.Context) error {
	userID, _, err := h.UserFromContext(c)
	if userID == nil || err != nil {
		return err
	}

	resp, err := h.eventSvc.ListJoinedEvents(c.Request().Context(), *userID)
	if err != nil {
		h.logger.Error("error on list joined events", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, resp)
}

// ListCategories godoc
// @Summary List categories
// @Description Returns root categories from the category table
// @Tags categories
// @Produce json
// @Success 200 {object} dto.CategoriesResponse
// @Failure 500 {object} dto.Msg
// @Router /api/v1/user/category [get]
func (h *EventHandler) ListCategories(c *echo.Context) error {
	resp, err := h.eventSvc.ListCategories(c.Request().Context())
	if err != nil {
		h.logger.Error("error on list categories", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateEvent godoc
// @Summary Update event
// @Description Updates event fields except created_at; sets updated_at on server
// @Tags events
// @Accept json
// @Produce json
// @Param event body dto.UpdateEventRequest true "Event payload (id in body)"
// @Success 200 {object} dto.Msg
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event [put]
func (h *EventHandler) UpdateEvent(c *echo.Context) error {
	userID, role, err := h.UserFromContext(c)
	if userID == nil || role == "" || err != nil {
		return err
	}

	var req dto.UpdateEventRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on update event: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if req.ID < 0 {
		h.logger.Info("error on update event: invalid event id", zap.Int("id", req.ID))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "invalid event id"})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on update event: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	// проверка может ли пользователь редактировать данное событие
	if status, err := h.checkPerms(c, req.ID, *userID, role); !status {
		return err
	}

	err = h.eventSvc.UpdateEvent(c.Request().Context(), &req)
	if errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on update event: reference not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on update event: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrEventAlreadyPassed) {
		h.logger.Info("error on update event: event time is in the past", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on update event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.Msg{Msg: MsgOK})
}

// JoinEvent godoc
// @Summary Join event
// @Description Adds current user as event participant
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.Msg
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 409 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id}/join [post]
func (h *EventHandler) JoinEvent(c *echo.Context) error {
	userID, _, err := h.UserFromContext(c)
	if userID == nil || err != nil {
		return err
	}

	var req dto.EventIDRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on join event: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on join event: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	err = h.eventSvc.JoinEvent(c.Request().Context(), req.ID, *userID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on join event: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on join event: user not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		h.logger.Info("error on join event: already joined", zap.Error(err))
		return c.JSON(http.StatusConflict, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrEventAlreadyPassed) {
		h.logger.Info("error on join event: event already passed", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on join event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.Msg{Msg: MsgOK})
}

// LeaveEvent godoc
// @Summary Leave event
// @Description Removes current user from event participants
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.Msg
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id}/join [delete]
func (h *EventHandler) LeaveEvent(c *echo.Context) error {
	userID, _, err := h.UserFromContext(c)
	if userID == nil || err != nil {
		return err
	}

	var req dto.EventIDRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on leave event: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on leave event: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	err = h.eventSvc.LeaveEvent(c.Request().Context(), req.ID, *userID)
	if errors.Is(err, exceptions.ErrNotFound) || errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on leave event: not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on leave event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.Msg{Msg: MsgOK})
}

// UploadEventImage godoc
// @Summary Upload event image
// @Description Uploads or replaces image for an event
// @Tags events
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Event ID"
// @Param image formData file true "Image file"
// @Success 200 {object} dto.ImageIDResponse
// @Failure 400 {object} dto.Msg
// @Failure 403 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id}/image [put]
func (h *EventHandler) UploadEventImage(c *echo.Context) error {
	userID, role, err := h.UserFromContext(c)
	if userID == nil || role == "" || err != nil {
		return err
	}

	var req dto.EventIDRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on upload event image: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on upload event image: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	if status, err := h.checkPerms(c, req.ID, *userID, role); !status {
		return err
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		fileHeader, err = c.FormFile("file")
	}
	if err != nil {
		h.logger.Info("error on upload event image: no file", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Info("error on upload event image: open file", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: err.Error()})
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("error on upload event image: read file", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	imageID, err := h.eventSvc.UploadEventImage(c.Request().Context(), req.ID, fileHeader.Filename, data)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on upload event image: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrEventAlreadyPassed) {
		h.logger.Info("error on upload event image: event already passed", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on upload event image", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.ImageIDResponse{ID: imageID})
}

// AddEventCategory godoc
// @Summary Add event category
// @Description Adds existing category to event
// @Tags events
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param category body dto.AddEventCategoryRequest true "Category payload"
// @Success 200 {object} dto.Msg
// @Failure 400 {object} dto.Msg
// @Failure 403 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 409 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id}/category [post]
func (h *EventHandler) AddEventCategory(c *echo.Context) error {
	userID, role, err := h.UserFromContext(c)
	if userID == nil || role == "" || err != nil {
		return err
	}

	var req dto.AddEventCategoryRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on add event category: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on add event category: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	if status, err := h.checkPerms(c, req.EventID, *userID, role); !status {
		return err
	}

	err = h.eventSvc.AddEventCategory(c.Request().Context(), &req)
	if errors.Is(err, exceptions.ErrNotFound) || errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on add event category: reference not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		h.logger.Info("error on add event category: already added", zap.Error(err))
		return c.JSON(http.StatusConflict, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrEventAlreadyPassed) {
		h.logger.Info("error on add event category: event already passed", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on add event category", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.Msg{Msg: MsgOK})
}

// DeleteEvent godoc
// @Summary Delete event
// @Description Soft-deletes event (deleted = true)
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.Msg
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 409 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id} [delete]
func (h *EventHandler) DeleteEvent(c *echo.Context) error {
	userID, role, err := h.UserFromContext(c)
	if userID == nil || role == "" || err != nil {
		return err
	}

	var req dto.EventIDRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on bind", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on validate", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	// проверка может ли пользователь редактировать данное событие
	if status, err := h.checkPerms(c, req.ID, *userID, role); !status {
		return err
	}

	err = h.eventSvc.DeleteEvent(c.Request().Context(), req.ID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("event not found on delete", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrAlreadyDeleted) {
		h.logger.Info("event already deleted", zap.Error(err))
		return c.JSON(http.StatusConflict, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on delete event", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, dto.Msg{Msg: MsgOK})
}

func (h *EventHandler) UserFromContext(c *echo.Context) (*int, string, error) {
	userID := utils.GetUserID(c)
	role := utils.GetUserRole(c)
	if userID == nil || role == "" {
		h.logger.Error("unexpected context state: no user in echo.Context", zap.Error(exceptions.ErrNoUserID), zap.Intp("user_id", userID), zap.String("role", role))
		return nil, "", c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}
	return userID, role, nil
}

func (h *EventHandler) checkPerms(c *echo.Context, eventID int, userID int, role string) (bool, error) {
	access, err := h.userSvc.CheckPerms(c.Request().Context(), eventID, userID, role)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on check permissions: event not found", zap.Error(err))
		return false, c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on check permissions: some error", zap.Error(err))
		return false, c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}
	if !access {
		h.logger.Info("permission denied", zap.Int("event_id", eventID), zap.Int("user_id", userID), zap.String("role", role))
		return false, c.JSON(http.StatusForbidden, dto.Msg{Msg: MsgPermissionDenied})
	}

	return true, nil
}

func uniquePositiveIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
