package handlers

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"JoinUp/internal/utils"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	ctx       context.Context
	logger    *zap.Logger
	reviewSvc ReviewService
}

func NewReviewHandler(ctx context.Context, logger *zap.Logger, reviewSvc ReviewService) ReviewHandler {
	return ReviewHandler{ctx: ctx, logger: logger, reviewSvc: reviewSvc}
}

// CreateReview godoc
// @Summary Create review
// @Description Creates a review for a past event by the current user
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param review body dto.CreateReviewRequest true "Review payload"
// @Success 201 {object} dto.ReviewResponse
// @Failure 400 {object} dto.Msg
// @Failure 401 {object} dto.Msg
// @Failure 403 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 409 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/user/event/{id}/review [post]
func (h *ReviewHandler) CreateReview(c *echo.Context) error {
	userID := utils.GetUserID(c)
	if userID == nil {
		h.logger.Error("unexpected context state: no user in echo.Context", zap.Error(exceptions.ErrNoUserID))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	var req dto.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on create review: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on create review: validate error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getValidationMsg(err)})
	}

	resp, err := h.reviewSvc.CreateReview(c.Request().Context(), req.EventID, *userID, req.Rating, req.Text)
	if errors.Is(err, exceptions.ErrInvalidPayload) {
		h.logger.Info("error on create review: invalid payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on create review: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrEventNotPassed) {
		h.logger.Info("error on create review: event not passed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrNotParticipant) {
		h.logger.Info("error on create review: user is not a participant", zap.Error(err))
		return c.JSON(http.StatusForbidden, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrOrganizerCannotReview) {
		h.logger.Info("error on create review: organizer cannot review own event", zap.Error(err))
		return c.JSON(http.StatusForbidden, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		h.logger.Info("error on create review: review already exists", zap.Error(err))
		return c.JSON(http.StatusConflict, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on create review", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusCreated, resp)
}

// ListReviews godoc
// @Summary List event reviews
// @Description Returns all reviews for a given event ordered by created_at desc
// @Tags reviews
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.ReviewsResponse
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/event/{id}/review [get]
func (h *ReviewHandler) ListReviews(c *echo.Context) error {
	var req dto.ReviewIDRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on list reviews: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on list reviews: validate error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getValidationMsg(err)})
	}

	resp, err := h.reviewSvc.ListReviews(c.Request().Context(), req.EventID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on list reviews: event not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on list reviews", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, resp)
}
