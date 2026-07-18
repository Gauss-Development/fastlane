package errors

import "net/http"

type OrderError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *OrderError) Error() string { return e.Message }

func newErr(code, message string, statusCode int) *OrderError {
	return &OrderError{Code: code, Message: message, StatusCode: statusCode}
}

var (
	ErrOrderNotFound      = newErr("ORDER_NOT_FOUND", "Order not found", http.StatusNotFound)
	ErrOrderExists        = newErr("ORDER_EXISTS", "Order already exists for this quote", http.StatusConflict)
	ErrUnauthorizedAccess = newErr("UNAUTHORIZED_ACCESS", "You don't have permission to perform this action", http.StatusForbidden)
	ErrInvalidRequest     = newErr("INVALID_REQUEST", "Invalid request parameters", http.StatusBadRequest)
	ErrInvalidTransition  = newErr("INVALID_TRANSITION", "State transition not allowed", http.StatusUnprocessableEntity)
	ErrServiceUnavailable = newErr("SERVICE_UNAVAILABLE", "Order service temporarily unavailable", http.StatusServiceUnavailable)
)
