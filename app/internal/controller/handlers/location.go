package handlers

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

func apiVersion(c *echo.Context) string {
	version, _, _ := strings.Cut(strings.TrimPrefix(c.Path(), "/api/"), "/")
	return version
}

func resourceLocation(c *echo.Context, resource string, id int) string {
	return "/api/" + apiVersion(c) + "/" + resource + "/" + strconv.Itoa(id)
}
