package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api-gateway/pkg/logger"
	"api-gateway/pkg/utils"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
)

const defaultSearchLimit = 20

// SearchClient is the minimal interface used by SearchHandler for testability.
type SearchClient interface {
	Search(ctx context.Context, query, requestingUserID string, limit int32, overrides *searchv1.ParsedSpecs) (*searchv1.SearchResponse, error)
}

type SearchHandler struct {
	searchClient SearchClient
	logger       *logger.Logger
}

func NewSearchHandler(searchClient SearchClient, logger *logger.Logger) *SearchHandler {
	return &SearchHandler{searchClient: searchClient, logger: logger}
}

// specChips mirrors search.v1.ParsedSpecs over JSON. The results screen sends
// it back as spec_overrides when a buyer edits a chip (GAU-247).
type specChips struct {
	DataRate      string   `json:"data_rate"`
	FormFactor    string   `json:"form_factor"`
	ReachKm       float64  `json:"reach_km"`
	WavelengthNm  int32    `json:"wavelength_nm"`
	Compatibility []string `json:"compatibility"`
	FiberType     string   `json:"fiber_type"`
	QtyEstimated  int32    `json:"qty_estimated"`
	FreeText      string   `json:"free_text"`
}

type searchRequestBody struct {
	Query         string     `json:"query"`
	Limit         int32      `json:"limit"`
	SpecOverrides *specChips `json:"spec_overrides"`
}

// Search proxies POST /api/v1/search to the search service's hybrid AI pipeline.
func (h *SearchHandler) Search(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	requestingUserID := userID.(string)

	var body searchRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_QUERY", "Search query is required")
		return
	}

	limit := int32(defaultSearchLimit)
	if body.Limit > 0 {
		limit = body.Limit
	}

	resp, err := h.searchClient.Search(c.Request.Context(), query, requestingUserID, limit, chipsToProto(body.SpecOverrides))
	if err != nil {
		h.handleSearchError(c, err)
		return
	}

	data := map[string]interface{}{
		"parsed_specs": parsedSpecsToMap(resp.GetParsedSpecs()),
		"results":      productHitsToMap(resp.GetResults()),
		"query_id":     resp.GetQueryId(),
	}
	utils.SuccessResponse(c, http.StatusOK, "Search completed successfully", data)
}

func chipsToProto(s *specChips) *searchv1.ParsedSpecs {
	if s == nil {
		return nil
	}
	return &searchv1.ParsedSpecs{
		DataRate:      s.DataRate,
		FormFactor:    s.FormFactor,
		ReachKm:       s.ReachKm,
		WavelengthNm:  s.WavelengthNm,
		Compatibility: s.Compatibility,
		FiberType:     s.FiberType,
		QtyEstimated:  s.QtyEstimated,
		FreeText:      s.FreeText,
	}
}

func parsedSpecsToMap(s *searchv1.ParsedSpecs) map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"data_rate":     s.GetDataRate(),
		"form_factor":   s.GetFormFactor(),
		"reach_km":      s.GetReachKm(),
		"wavelength_nm": s.GetWavelengthNm(),
		"compatibility": s.GetCompatibility(),
		"fiber_type":    s.GetFiberType(),
		"qty_estimated": s.GetQtyEstimated(),
		"free_text":     s.GetFreeText(),
	}
}

func productHitsToMap(hits []*searchv1.ProductHit) []map[string]interface{} {
	if hits == nil {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(hits))
	for _, p := range hits {
		var specs map[string]interface{}
		if s := p.GetSpecs(); s != nil {
			specs = s.AsMap()
		}
		out = append(out, map[string]interface{}{
			"id":                p.GetId(),
			"supplier_id":       p.GetSupplierId(),
			"sku":               p.GetSku(),
			"name":              p.GetName(),
			"name_zh":           p.GetNameZh(),
			"category":          p.GetCategory(),
			"specs":             specs,
			"price_usd":         p.GetPriceUsd(),
			"moq":               p.GetMoq(),
			"stock_qty":         p.GetStockQty(),
			"lead_time_days":    p.GetLeadTimeDays(),
			"datasheet_url":     p.GetDatasheetUrl(),
			"supplier_name":     p.GetSupplierName(),
			"supplier_city":     p.GetSupplierCity(),
			"supplier_verified": p.GetSupplierVerified(),
			"match_score":       p.GetMatchScore(),
			"vector_distance":   p.GetVectorDistance(),
			"match_explanation": p.GetMatchExplanation(),
		})
	}
	return out
}

func (h *SearchHandler) handleSearchError(c *gin.Context, err error) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			return
		case codes.Unavailable:
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "SEARCH_UNAVAILABLE", "Search service temporarily unavailable")
			return
		}
	}
	h.logger.Error("Search failed: " + err.Error())
	utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_FAILED", "Failed to search")
}
