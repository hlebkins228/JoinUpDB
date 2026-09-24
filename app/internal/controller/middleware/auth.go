package middleware

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/controller/handlers"
	"JoinUp/internal/settings"
	"JoinUp/internal/utils/jwt"
	"errors"
	"net/http"

	JWT "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func JWTAuthorization(logger *zap.Logger, mgr *jwt.JwtManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, err := mgr.ValidateTokenFromHeader(c.Request().Header)
			if err != nil {
				var msg string
				if errors.Is(err, JWT.ErrTokenExpired) {
					msg = handlers.MsgTokenExpired
				} else {
					msg = handlers.MsgAuthError
				}
				logger.Info(msg, zap.Error(err))
				return c.JSON(http.StatusUnauthorized, dto.Msg{Msg: msg})
			}
			c.Set(settings.ContextKeyUserID, claims.UserID)
			c.Set(settings.ContextKeyUserRole, claims.Role)
			return next(c)
		}
	}
}
