package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents application errors
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NotFound creates a 404 error
func NotFound(message string) *AppError {
	return NewAppError(http.StatusNotFound, message, nil)
}

// Unauthorized creates a 401 error
func Unauthorized(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, message, nil)
}

// BadRequest creates a 400 error
func BadRequest(message string) *AppError {
	return NewAppError(http.StatusBadRequest, message, nil)
}

// Internal creates a 500 error
func Internal(message string, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, message, err)
}

// ErrorResponse handles errors in Gin context
func ErrorResponse(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, gin.H{"error": appErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
