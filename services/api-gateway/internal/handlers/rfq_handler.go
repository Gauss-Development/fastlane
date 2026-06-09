package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"api-gateway/pkg/logger"
	"api-gateway/pkg/utils"

	authv1 "github.com/nikitashilov/microblog_grpc/proto/auth/v1"
	rfqv1 "github.com/nikitashilov/microblog_grpc/proto/rfq/v1"
)

// RFQClientAPI is the minimal client surface used by RFQHandler (testable).
type RFQClientAPI interface {
	CreateRFQ(ctx context.Context, rfq *rfqv1.RFQ) (*rfqv1.RFQ, error)
	GetRFQ(ctx context.Context, id, requestingUserID string) (*rfqv1.RFQ, error)
	ListRFQs(ctx context.Context, buyerID, status string, limit, offset int32) (*rfqv1.ListRFQsResponse, error)
	ListQuotesForRFQ(ctx context.Context, rfqID, requestingUserID string) (*rfqv1.ListQuotesResponse, error)
	GetRFQForSupplier(ctx context.Context, rfqID, supplierID string) (*rfqv1.SupplierRFQView, error)
	AddQuote(ctx context.Context, quote *rfqv1.Quote) (*rfqv1.Quote, error)
}

// MagicLinkValidator is the slice of AuthClient the supplier endpoints need.
type MagicLinkValidator interface {
	ValidateMagicLinkToken(ctx context.Context, token string) (*authv1.ValidateMagicLinkTokenResponse, error)
}

type RFQHandler struct {
	rfqClient  RFQClientAPI
	magicLinks MagicLinkValidator
	logger     *logger.Logger
}

func NewRFQHandler(rfqClient RFQClientAPI, magicLinks MagicLinkValidator, logger *logger.Logger) *RFQHandler {
	return &RFQHandler{rfqClient: rfqClient, magicLinks: magicLinks, logger: logger}
}

type createRFQBody struct {
	QueryText         string                 `json:"query_text"`
	ParsedSpecs       map[string]interface{} `json:"parsed_specs"`
	MatchedProductIDs []string               `json:"matched_product_ids"`
	Qty               int32                  `json:"qty"`
	TargetDate        string                 `json:"target_date"`
	ShippingAddress   string                 `json:"shipping_address"`
	Notes             string                 `json:"notes"`
	BuyerCompany      string                 `json:"buyer_company"`
}

// CreateRFQ handles POST /api/v1/rfqs.
func (h *RFQHandler) CreateRFQ(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	userEmail, _ := c.Get("userEmail")
	buyerEmail, _ := userEmail.(string)

	var body createRFQBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if strings.TrimSpace(body.QueryText) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_QUERY", "query_text is required")
		return
	}
	if len(body.MatchedProductIDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_PRODUCTS", "matched_product_ids must reference at least one product")
		return
	}

	buyerCompany := strings.TrimSpace(body.BuyerCompany)
	if buyerCompany == "" {
		buyerCompany = companyFromEmail(buyerEmail)
	}

	var specs *structpb.Struct
	if body.ParsedSpecs != nil {
		if s, err := structpb.NewStruct(body.ParsedSpecs); err == nil {
			specs = s
		}
	}

	rfq, err := h.rfqClient.CreateRFQ(c.Request.Context(), &rfqv1.RFQ{
		BuyerId:           userID.(string),
		BuyerEmail:        buyerEmail,
		BuyerCompany:      buyerCompany,
		QueryText:         strings.TrimSpace(body.QueryText),
		ParsedSpecs:       specs,
		MatchedProductIds: body.MatchedProductIDs,
		Qty:               body.Qty,
		TargetDate:        body.TargetDate,
		ShippingAddress:   body.ShippingAddress,
		Notes:             body.Notes,
	})
	if err != nil {
		h.handleRFQError(c, err, "CREATE_RFQ_FAILED", "Failed to create RFQ")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "RFQ created successfully", rfqToMap(rfq, true))
}

// GetRFQ handles GET /api/v1/rfqs/:id.
func (h *RFQHandler) GetRFQ(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	rfq, err := h.rfqClient.GetRFQ(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleRFQError(c, err, "GET_RFQ_FAILED", "Failed to retrieve RFQ")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "RFQ retrieved successfully", rfqToMap(rfq, true))
}

// ListRFQs handles GET /api/v1/rfqs.
func (h *RFQHandler) ListRFQs(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit := parseQueryInt(c, "limit", 20, 1, 100)
	offset := parseQueryInt(c, "offset", 0, 0, 1<<30)

	resp, err := h.rfqClient.ListRFQs(c.Request.Context(), userID.(string), c.Query("status"), limit, offset)
	if err != nil {
		h.handleRFQError(c, err, "LIST_RFQS_FAILED", "Failed to list RFQs")
		return
	}

	rfqs := make([]map[string]interface{}, 0, len(resp.GetRfqs()))
	for _, rfq := range resp.GetRfqs() {
		rfqs = append(rfqs, rfqToMap(rfq, true))
	}
	utils.SuccessResponse(c, http.StatusOK, "RFQs retrieved successfully", map[string]interface{}{
		"rfqs":   rfqs,
		"total":  resp.GetTotal(),
		"limit":  limit,
		"offset": offset,
	})
}

// ListQuotes handles GET /api/v1/rfqs/:id/quotes.
func (h *RFQHandler) ListQuotes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	resp, err := h.rfqClient.ListQuotesForRFQ(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleRFQError(c, err, "LIST_QUOTES_FAILED", "Failed to list quotes")
		return
	}

	quotes := make([]map[string]interface{}, 0, len(resp.GetQuotes()))
	for _, quote := range resp.GetQuotes() {
		quotes = append(quotes, quoteToMap(quote))
	}
	utils.SuccessResponse(c, http.StatusOK, "Quotes retrieved successfully", map[string]interface{}{
		"quotes": quotes,
	})
}

// SupplierGetRFQ handles GET /api/v1/supplier-rfq/:token — public, gated only
// by the signed magic-link token.
func (h *RFQHandler) SupplierGetRFQ(c *gin.Context) {
	scope, ok := h.resolveMagicLink(c)
	if !ok {
		return
	}

	view, err := h.rfqClient.GetRFQForSupplier(c.Request.Context(), scope.GetRfqId(), scope.GetSupplierId())
	if err != nil {
		h.handleRFQError(c, err, "GET_SUPPLIER_RFQ_FAILED", "Failed to load RFQ")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "RFQ retrieved successfully", map[string]interface{}{
		// buyer_email is blanked by the service for supplier-facing views.
		"rfq":           rfqToMap(view.GetRfq(), false),
		"quote":         quoteToMap(view.GetQuote()),
		"supplier_name": view.GetSupplierName(),
		"supplier_id":   scope.GetSupplierId(),
	})
}

type submitQuoteBody struct {
	PriceUSD     float64 `json:"price_usd"`
	LeadTimeDays int32   `json:"lead_time_days"`
	ValidityDate string  `json:"validity_date"`
	Notes        string  `json:"notes"`
}

// SupplierSubmitQuote handles POST /api/v1/supplier-rfq/:token/quote.
func (h *RFQHandler) SupplierSubmitQuote(c *gin.Context) {
	scope, ok := h.resolveMagicLink(c)
	if !ok {
		return
	}

	var body submitQuoteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if body.PriceUSD <= 0 || body.LeadTimeDays <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_QUOTE", "price_usd and lead_time_days must be positive")
		return
	}

	quote, err := h.rfqClient.AddQuote(c.Request.Context(), &rfqv1.Quote{
		RfqId:         scope.GetRfqId(),
		SupplierId:    scope.GetSupplierId(),
		PriceUsd:      body.PriceUSD,
		LeadTimeDays:  body.LeadTimeDays,
		ValidityDate:  body.ValidityDate,
		SupplierNotes: body.Notes,
	})
	if err != nil {
		h.handleRFQError(c, err, "SUBMIT_QUOTE_FAILED", "Failed to submit quote")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Quote submitted successfully", quoteToMap(quote))
}

func (h *RFQHandler) resolveMagicLink(c *gin.Context) (*authv1.ValidateMagicLinkTokenResponse, bool) {
	token := c.Param("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_TOKEN", "Magic link token is required")
		return nil, false
	}

	resp, err := h.magicLinks.ValidateMagicLinkToken(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("magic link validation failed: " + err.Error())
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Could not validate the link, try again")
		return nil, false
	}
	if !resp.GetValid() {
		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_MAGIC_LINK", "This link is invalid or has expired")
		return nil, false
	}
	return resp, true
}

func (h *RFQHandler) handleRFQError(c *gin.Context, err error, fallbackCode, fallbackMessage string) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			return
		case codes.NotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", st.Message())
			return
		case codes.AlreadyExists:
			utils.ErrorResponse(c, http.StatusConflict, "ALREADY_EXISTS", st.Message())
			return
		case codes.PermissionDenied:
			utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", st.Message())
			return
		case codes.Unauthenticated:
			utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", st.Message())
			return
		case codes.Unavailable:
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "RFQ_UNAVAILABLE", "RFQ service temporarily unavailable")
			return
		}
	}
	h.logger.Error(fallbackMessage + ": " + err.Error())
	utils.ErrorResponse(c, http.StatusInternalServerError, fallbackCode, fallbackMessage)
}

// rfqToMap serializes an RFQ; includeBuyer guards buyer PII on
// supplier-facing responses.
func rfqToMap(rfq *rfqv1.RFQ, includeBuyer bool) map[string]interface{} {
	if rfq == nil {
		return nil
	}
	var specs map[string]interface{}
	if s := rfq.GetParsedSpecs(); s != nil {
		specs = s.AsMap()
	}
	out := map[string]interface{}{
		"id":                  rfq.GetId(),
		"query_text":          rfq.GetQueryText(),
		"parsed_specs":        specs,
		"matched_product_ids": rfq.GetMatchedProductIds(),
		"status":              rfq.GetStatus(),
		"qty":                 rfq.GetQty(),
		"target_date":         rfq.GetTargetDate(),
		"shipping_address":    rfq.GetShippingAddress(),
		"notes":               rfq.GetNotes(),
		"buyer_company":       rfq.GetBuyerCompany(),
		"created_at":          timestampString(rfq.GetCreatedAt()),
	}
	if includeBuyer {
		out["buyer_id"] = rfq.GetBuyerId()
		out["buyer_email"] = rfq.GetBuyerEmail()
	}
	return out
}

func quoteToMap(quote *rfqv1.Quote) map[string]interface{} {
	if quote == nil {
		return nil
	}
	return map[string]interface{}{
		"id":             quote.GetId(),
		"rfq_id":         quote.GetRfqId(),
		"supplier_id":    quote.GetSupplierId(),
		"product_id":     quote.GetProductId(),
		"price_usd":      quote.GetPriceUsd(),
		"lead_time_days": quote.GetLeadTimeDays(),
		"validity_date":  quote.GetValidityDate(),
		"supplier_notes": quote.GetSupplierNotes(),
		"match_score":    quote.GetMatchScore(),
		"status":         quote.GetStatus(),
		"submitted_at":   timestampString(quote.GetSubmittedAt()),
		"created_at":     timestampString(quote.GetCreatedAt()),
	}
}

func timestampString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

func companyFromEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return ""
	}
	domain := email[at+1:]
	if dot := strings.Index(domain, "."); dot > 0 {
		domain = domain[:dot]
	}
	return strings.ToUpper(domain)
}

func parseQueryInt(c *gin.Context, name string, def, min, max int32) int32 {
	raw := c.Query(name)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	value := int32(v)
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
