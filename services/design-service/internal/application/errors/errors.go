package errors

import "net/http"

type DesignError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *DesignError) Error() string {
	return e.Message
}

func newErr(code, message string, statusCode int) *DesignError {
	return &DesignError{Code: code, Message: message, StatusCode: statusCode}
}

var (
	ErrProjectNotFound    = newErr("PROJECT_NOT_FOUND", "Project not found", http.StatusNotFound)
	ErrFileNotFound       = newErr("FILE_NOT_FOUND", "File not found", http.StatusNotFound)
	ErrNDANotFound        = newErr("NDA_NOT_FOUND", "NDA not found", http.StatusNotFound)
	ErrUnauthorizedAccess = newErr("UNAUTHORIZED_ACCESS", "You don't have permission to access this resource", http.StatusForbidden)
	ErrInvalidRequest     = newErr("INVALID_REQUEST", "Invalid request parameters", http.StatusBadRequest)
	ErrUploadIncomplete   = newErr("UPLOAD_INCOMPLETE", "Uploaded object not found in storage or size mismatch; the upload did not complete", http.StatusBadRequest)
	ErrServiceUnavailable = newErr("SERVICE_UNAVAILABLE", "Design service temporarily unavailable", http.StatusServiceUnavailable)
)
