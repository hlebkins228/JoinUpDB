package handlers

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"JoinUp/internal/settings"
	"JoinUp/internal/utils"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type UserHandler struct {
	ctx     context.Context
	logger  *zap.Logger
	userSvc UserService
}

func NewUserHandler(ctx context.Context, logger *zap.Logger, userSvc UserService) UserHandler {
	return UserHandler{ctx: ctx, logger: logger, userSvc: userSvc}
}

// CreateUser godoc
// @Summary Create user
// @Description Creates new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User payload"
// @Success 201 {object} dto.UserIDResponse
// @Failure 400 {object} dto.Msg
// @Failure 409 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *echo.Context) error {
	var req dto.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on create user: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on create user: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	id, err := h.userSvc.CreateUser(c.Request().Context(), &req)
	if errors.Is(err, exceptions.ErrAlreadyExists) {
		h.logger.Info("error on create user: login already exists", zap.Error(err))
		return c.JSON(http.StatusConflict, dto.Msg{Msg: err.Error()})
	}
	if errors.Is(err, exceptions.ErrNotExists) {
		h.logger.Info("error on create user: reference not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on create user", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	c.Response().Header().Set("Location", "/api/v1/users/"+strconv.Itoa(id))
	return c.JSON(http.StatusCreated, dto.UserIDResponse{ID: id})
}

// GetUser godoc
// @Summary Get user by id
// @Description Returns user by identifier
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *echo.Context) error {
	var req dto.UserIDRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on get user: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on get user: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	user, err := h.userSvc.GetUser(c.Request().Context(), req.ID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on get user: user not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on get user", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update current user
// @Description Updates current user's editable fields
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UpdateUserRequest true "User payload"
// @Success 204
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	userID := utils.GetUserID(c)
	if userID == nil {
		h.logger.Error("unexpected context state: no user in echo.Context", zap.Error(exceptions.ErrNoUserID))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on update user: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on update user: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	err := h.userSvc.UpdateUser(c.Request().Context(), *userID, &req)
	if errors.Is(err, exceptions.ErrNotExists) || errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on update user: reference not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on update user", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	return c.NoContent(http.StatusNoContent)
}

// UploadImage godoc
// @Summary Upload image
// @Description Uploads image file and returns image id
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file"
// @Success 201 {object} dto.ImageIDResponse
// @Failure 400 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Security BearerAuth
// @Router /api/v1/images [post]
func (h *UserHandler) UploadImage(c *echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		fileHeader, err = c.FormFile("file")
	}
	if err != nil {
		h.logger.Info("error on upload image: no file", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Info("error on upload image: open file", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: err.Error()})
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("error on upload image: read file", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	id, err := h.userSvc.UploadImage(c.Request().Context(), fileHeader.Filename, data)
	if err != nil {
		h.logger.Error("error on upload image", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	c.Response().Header().Set("Location", "/api/v1/images/"+strconv.Itoa(id))
	return c.JSON(http.StatusCreated, dto.ImageIDResponse{ID: id})
}

// GetImage godoc
// @Summary Get image by id
// @Description Returns raw image bytes
// @Tags images
// @Produce octet-stream
// @Param id path int true "Image ID"
// @Success 200 {file} binary
// @Failure 400 {object} dto.Msg
// @Failure 404 {object} dto.Msg
// @Failure 422 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/images/{id} [get]
func (h *UserHandler) GetImage(c *echo.Context) error {
	var req dto.UserIDRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("error on get image: bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	if err := c.Validate(&req); err != nil {
		h.logger.Info("error on get image: validate error", zap.Error(err))
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}

	image, err := h.userSvc.GetImage(c.Request().Context(), req.ID)
	if errors.Is(err, exceptions.ErrNotFound) {
		h.logger.Info("error on get image: image not found", zap.Error(err))
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on get image", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	contentType := http.DetectContentType(image.Data)
	return c.Blob(http.StatusOK, contentType, image.Data)
}

// Auth godoc
// @Summary Authorize user
// @Description Returns JWT token for login/password headers or credentials in the JSON body
// @Tags auth
// @Produce json
// @Accept json
// @Param login header string false "Login (requires password header)"
// @Param password header string false "Password (requires login header)"
// @Param credentials body dto.AuthRequest false "Credentials when login/password headers are absent"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.Msg
// @Failure 500 {object} dto.Msg
// @Router /api/v1/tokens [post]
func (h *UserHandler) Auth(c *echo.Context) error {
	req := dto.AuthRequest{
		Login:    c.Request().Header.Get("login"),
		Password: c.Request().Header.Get("password"),
	}
	if req.Login == "" && req.Password == "" {
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
		}
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: getValidationMsg(err)})
	}
	c.Response().Header().Set("Cache-Control", "no-store")

	token, err := h.userSvc.Auth(c.Request().Context(), req.Login, req.Password)
	if errors.Is(err, exceptions.ErrInvalidCredentials) {
		h.logger.Info("error on auth: invalid credentials", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		h.logger.Error("error on auth", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}

	c.Response().Header().Set(settings.AuthHeader, "Bearer "+token)
	return c.JSON(http.StatusOK, dto.AuthResponse{Token: token})
}
