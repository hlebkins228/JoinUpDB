package utils

import (
	"JoinUp/internal/settings"

	"github.com/labstack/echo/v5"
)

func GetUserID(c *echo.Context) *int {
	userID, ok := c.Get(settings.ContextKeyUserID).(int)
	if !ok {
		return nil
	}

	return &userID
}

func GetUserRole(c *echo.Context) string {
	role, ok := c.Get(settings.ContextKeyUserRole).(string)
	if !ok {
		return ""
	}

	return role
}

