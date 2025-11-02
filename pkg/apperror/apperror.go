package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrPermission   = errors.New("permission denied")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal server error")
	ErrUnauthorized = errors.New("unauthorized")
)

type AppError struct {
	BaseError error
	Message   string
	Details   string
	Err       error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (Details: %s, Cause: %v)", e.BaseError.Error(), e.Message, e.Details, e.Err)
	}
	return fmt.Sprintf("%s: %s (Details: %s)", e.BaseError.Error(), e.Message, e.Details)
}

func (e *AppError) Unwrap() error {
	return e.BaseError
}

func NewAppError(base error, msg, details string, err error) *AppError {
	return &AppError{BaseError: base, Message: msg, Details: details, Err: err}
}

func NewNotFound(resource, identifier string) *AppError {
	msg := fmt.Sprintf("%s not found", resource)
	details := fmt.Sprintf("%s with identifier '%s' was not found", resource, identifier)
	return NewAppError(ErrNotFound, msg, details, nil)
}

func NewInvalidInput(details string, err error) *AppError {
	return NewAppError(ErrInvalidInput, "Invalid input provided", details, err)
}

func NewConflict(resource, field, value string) *AppError {
	msg := fmt.Sprintf("%s conflict", resource)
	details := fmt.Sprintf("%s with %s '%s' already exists", resource, field, value)
	return NewAppError(ErrConflict, msg, details, nil)
}

func NewInternal(details string, err error) *AppError {
	return NewAppError(ErrInternal, "An internal server error occurred", details, err)
}

func NewUnauthorized(details string, err error) *AppError {
	return NewAppError(ErrUnauthorized, "Invalid credentials", details, err)
}

func NewPermissionDenied(details string) *AppError {
	return NewAppError(ErrPermission, "Permission denied", details, nil)
}

func ToHTTPStatus(err error) int {
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrInvalidInput) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, ErrPermission) {
		return http.StatusForbidden
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func (e *AppError) ToJSON() gin.H {
	return gin.H{
		"error":   e.BaseError.Error(),
		"message": e.Message,
	}
}
