package errors

import "net/http"

// RFQ flow reuses the PostError envelope so the existing gRPC error mapping
// keeps working while the service is repurposed (post = inquiry/RFQ).
var (
	ErrRFQNotFound        = NewPostError("RFQ_NOT_FOUND", "RFQ not found", http.StatusNotFound)
	ErrQuoteNotFound      = NewPostError("QUOTE_NOT_FOUND", "Quote not found for this RFQ and supplier", http.StatusNotFound)
	ErrQuoteAlreadyExists = NewPostError("QUOTE_ALREADY_SUBMITTED", "Quote was already submitted for this RFQ", http.StatusConflict)
	ErrInvalidRFQData     = NewPostError("INVALID_RFQ_DATA", "Invalid RFQ data provided", http.StatusBadRequest)
	ErrNoMatchedProducts  = NewPostError("NO_MATCHED_PRODUCTS", "RFQ must reference at least one catalog product", http.StatusBadRequest)
	ErrRFQCreationFailed  = NewPostError("RFQ_CREATION_FAILED", "Failed to create RFQ", http.StatusInternalServerError)
)
