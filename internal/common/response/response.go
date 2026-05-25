package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleError inspects a domain error and outputs the correct standardized HTTP response envelope

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Default values for unknown errors
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	// Map known domain errors to their respective HTTP behaviors
	switch {
	case errors.Is(err, ErrNotFound):
		statusCode = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		message = err.Error()
	case errors.Is(err, ErrAlreadyExists):
		statusCode = http.StatusConflict
		message = err.Error()
	case errors.Is(err, ErrInvalidInput):
		statusCode = http.StatusBadRequest
		message = err.Error()
	}

	// Execute the centralized error payload delivery
	Error(c, statusCode, message, nil)
}

// JSONResponse defines the uniform format for every successful API response
type JSONResponse struct {
	Status  string      `json:"status"`            // "success"
	Message string      `json:"message,omitempty"` // Optional human-readable message
	Data    interface{} `json:"data,omitempty"`    // Dynamic payload envelope
}

// APIError defines the structural layout for error contexts
type APIError struct {
	Status  string      `json:"status"`            // "error"
	Code    int         `json:"code"`              // HTTP Status Code
	Message string      `json:"message"`           // High-level safe message
	Details interface{} `json:"details,omitempty"` // Validation specifics or structural logs
}

// Success writes a standardized 2xx success envelope to the Gin context
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, JSONResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error writes a standardized error signature to the client
func Error(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, APIError{
		Status:  "error",
		Code:    statusCode,
		Message: message,
		Details: details,
	})
}
