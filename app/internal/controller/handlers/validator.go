package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type Validator struct {
	Validator *validator.Validate
}

func (cv *Validator) Validate(i any) error {
	return cv.Validator.Struct(i)
}

func NewValidator() Validator {
	validate := validator.New()

	return Validator{Validator: validate}
}

func getValidationMsg(err error) string {
	if err == nil {
		return ""
	}
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	errs := make([]string, 0)
	for _, e := range validationErrors {
		errs = append(errs, fmt.Sprintf("field %s: expected %s, got `%s`", e.Field(), e.Tag(), e.Value()))
	}

	return strings.Join(errs, ";")
}

func getBindMsg(err error) string {
	if err == nil {
		return ""
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		if wrapped := errors.Unwrap(httpErr); wrapped != nil {
			return wrapped.Error()
		}
		return httpErr.Message
	}

	

	return err.Error()
}
