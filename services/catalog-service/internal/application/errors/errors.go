package errors

import "net/http"

type CatalogError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *CatalogError) Error() string { return e.Message }

func newErr(code, message string, statusCode int) *CatalogError {
	return &CatalogError{Code: code, Message: message, StatusCode: statusCode}
}

var (
	ErrManufacturerNotFound = newErr("MANUFACTURER_NOT_FOUND", "Manufacturer not found", http.StatusNotFound)
	ErrManufacturerExists   = newErr("MANUFACTURER_EXISTS", "Manufacturer profile already exists for this user", http.StatusConflict)
	ErrUnauthorizedAccess   = newErr("UNAUTHORIZED_ACCESS", "You don't have permission to perform this action", http.StatusForbidden)
	ErrInvalidRequest       = newErr("INVALID_REQUEST", "Invalid request parameters", http.StatusBadRequest)
	ErrServiceUnavailable   = newErr("SERVICE_UNAVAILABLE", "Catalog service temporarily unavailable", http.StatusServiceUnavailable)
)
