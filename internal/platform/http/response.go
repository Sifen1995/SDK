package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	statusCode := http.StatusInternalServerError
	message := "Internal server error"

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

	Error(c, statusCode, message, nil)
}

type JSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type APIError struct {
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, JSONResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, APIError{
		Status:  "error",
		Code:    statusCode,
		Message: message,
		Details: details,
	})
}

