package errors

import (
	"net/http"
)

type PostError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *PostError) Error() string {
	return e.Message
}

func NewPostError(code, message string, statusCode int) *PostError {
	return &PostError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Shared error envelope for the service (type kept as PostError for the gRPC
// error mapping in rfq_server.go; RFQ/quote errors live in rfq_errors.go).
var (
	ErrUnauthorizedAccess = NewPostError("UNAUTHORIZED_ACCESS", "You don't have permission to access this resource", http.StatusForbidden)
	ErrInvalidRequest     = NewPostError("INVALID_REQUEST", "Invalid request parameters", http.StatusBadRequest)
	ErrServiceUnavailable = NewPostError("SERVICE_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable)
)
