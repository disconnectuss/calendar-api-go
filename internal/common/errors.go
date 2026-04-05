package common

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/googleapi"
)

type APIError struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Errors     interface{} `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{StatusCode: statusCode, Message: message}
}

func NotFoundError(resource, id string) *APIError {
	return &APIError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("%s '%s' not found", resource, id),
	}
}

func ForbiddenError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

func UnauthorizedError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

func BadRequestError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func HandleGoogleAPIError(err error) *APIError {
	if apiErr, ok := err.(*googleapi.Error); ok {
		msg := apiErr.Message
		if msg == "" {
			msg = http.StatusText(apiErr.Code)
		}
		return &APIError{
			StatusCode: apiErr.Code,
			Message:    msg,
			Errors:     apiErr.Errors,
		}
	}
	return &APIError{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal server error",
	}
}

func RespondWithError(c *gin.Context, err error) {
	if apiErr, ok := err.(*APIError); ok {
		c.JSON(apiErr.StatusCode, apiErr)
		return
	}
	googleErr := HandleGoogleAPIError(err)
	c.JSON(googleErr.StatusCode, googleErr)
}
