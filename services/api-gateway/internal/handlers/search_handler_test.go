package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
)

type mockSearchClient struct {
	resp *searchv1.SearchResponse
	err  error
}

func (m *mockSearchClient) Search(ctx context.Context, query, requestingUserID string, limit int32, overrides *searchv1.ParsedSpecs) (*searchv1.SearchResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func TestSearchHandler_Search_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("info")
	mock := &mockSearchClient{
		resp: &searchv1.SearchResponse{
			ParsedSpecs: &searchv1.ParsedSpecs{DataRate: "100G", FormFactor: "QSFP28"},
			Results: []*searchv1.ProductHit{
				{Id: "p1", Sku: "QSFP28-100G-LR4", Name: "100G LR4", MatchScore: 92, MatchExplanation: "matches 100G + 10km"},
			},
			QueryId: "abc123",
		},
	}
	h := NewSearchHandler(mock, log)

	r := gin.New()
	r.POST("/search", func(c *gin.Context) {
		c.Set("userID", "requesting-user")
		h.Search(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(`{"query":"100G transceiver 10km"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "QSFP28-100G-LR4") {
		t.Errorf("expected result SKU in body, got %s", rec.Body.String())
	}
}

func TestSearchHandler_Search_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("info")
	h := NewSearchHandler(&mockSearchClient{}, log)

	r := gin.New()
	r.POST("/search", h.Search)

	req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(`{"query":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 when userID not set, got %d", rec.Code)
	}
}

func TestSearchHandler_Search_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("info")
	h := NewSearchHandler(&mockSearchClient{}, log)

	r := gin.New()
	r.POST("/search", func(c *gin.Context) {
		c.Set("userID", "user1")
		h.Search(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 when query missing, got %d", rec.Code)
	}
}
