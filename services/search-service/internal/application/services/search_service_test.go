package services

import (
	"context"
	"sync/atomic"
	"testing"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
	"search-service/internal/domain"
	"search-service/pkg/logger"
)

type fakeExtractor struct {
	specs   *searchv1.ParsedSpecs
	enabled bool
	calls   int32
}

func (f *fakeExtractor) ExtractSpecs(context.Context, string) (*searchv1.ParsedSpecs, error) {
	atomic.AddInt32(&f.calls, 1)
	return f.specs, nil
}
func (f *fakeExtractor) Enabled() bool { return f.enabled }

type fakeEmbedder struct {
	vec   []float32
	calls int32
}

func (f *fakeEmbedder) Embed(context.Context, []string, string) ([][]float32, error) {
	atomic.AddInt32(&f.calls, 1)
	return [][]float32{f.vec}, nil
}

type fakeCatalog struct{ hits []domain.CatalogHit }

func (f *fakeCatalog) VectorSearch(_ context.Context, _ []float32, limit int) ([]domain.CatalogHit, error) {
	if len(f.hits) > limit {
		return f.hits[:limit], nil
	}
	return f.hits, nil
}

type fakeExplainer struct {
	enabled bool
	calls   int32
}

func (f *fakeExplainer) Explain(_ context.Context, _ string, hit domain.CatalogHit) (string, error) {
	atomic.AddInt32(&f.calls, 1)
	return "matches because " + hit.SKU, nil
}
func (f *fakeExplainer) Enabled() bool { return f.enabled }

func newTestService(ext SpecExtractor, emb Embedder, cat CatalogRepo, exp Explainer) *SearchService {
	return NewSearchService(ext, emb, cat, exp, NoopCache{}, 0, logger.New("error"))
}

func hit(id, sku string, dist float64, specsJSON string) domain.CatalogHit {
	return domain.CatalogHit{ID: id, SKU: sku, VectorDistance: dist, SpecsJSON: []byte(specsJSON)}
}

func TestSearch_EmptyQuery(t *testing.T) {
	ext := &fakeExtractor{enabled: true}
	emb := &fakeEmbedder{vec: []float32{1}}
	s := newTestService(ext, emb, &fakeCatalog{}, &fakeExplainer{})
	resp, err := s.Search(context.Background(), &searchv1.SearchRequest{Query: "  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetResults()) != 0 {
		t.Errorf("expected no results for empty query, got %d", len(resp.GetResults()))
	}
	if atomic.LoadInt32(&emb.calls) != 0 {
		t.Error("embedder should not be called for empty query")
	}
}

func TestSearch_SpecFitBoostsRanking(t *testing.T) {
	// A is slightly closer by vector but wrong data rate; B is the exact match.
	cat := &fakeCatalog{hits: []domain.CatalogHit{
		hit("a", "SFP-10G-SR", 0.20, `{"data_rate":"10G"}`),
		hit("b", "QSFP28-100G-LR4", 0.25, `{"data_rate":"100G"}`),
	}}
	ext := &fakeExtractor{enabled: true, specs: &searchv1.ParsedSpecs{DataRate: "100G"}}
	exp := &fakeExplainer{enabled: true}
	s := newTestService(ext, &fakeEmbedder{vec: []float32{1}}, cat, exp)

	resp, err := s.Search(context.Background(), &searchv1.SearchRequest{Query: "100G transceiver", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetResults()) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.GetResults()))
	}
	if resp.GetResults()[0].GetSku() != "QSFP28-100G-LR4" {
		t.Errorf("spec-matching part should rank first, got %q", resp.GetResults()[0].GetSku())
	}
	if resp.GetParsedSpecs().GetDataRate() != "100G" {
		t.Errorf("parsed_specs not propagated: %v", resp.GetParsedSpecs())
	}
	if resp.GetQueryId() == "" {
		t.Error("query_id should be set")
	}
	for _, r := range resp.GetResults() {
		if r.GetMatchExplanation() == "" {
			t.Errorf("missing explanation for %q", r.GetSku())
		}
	}
}

func TestSearch_DegradedWithoutClaude(t *testing.T) {
	// Extractor + explainer disabled (no API key): vector-only, no specs/explanations.
	cat := &fakeCatalog{hits: []domain.CatalogHit{
		hit("a", "A", 0.40, `{}`),
		hit("b", "B", 0.10, `{}`),
	}}
	ext := &fakeExtractor{enabled: false}
	exp := &fakeExplainer{enabled: false}
	s := newTestService(ext, &fakeEmbedder{vec: []float32{1}}, cat, exp)

	resp, err := s.Search(context.Background(), &searchv1.SearchRequest{Query: "anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetParsedSpecs() != nil {
		t.Error("expected nil parsed_specs when extractor disabled")
	}
	if resp.GetResults()[0].GetSku() != "B" {
		t.Errorf("closest-by-vector should rank first, got %q", resp.GetResults()[0].GetSku())
	}
	if atomic.LoadInt32(&ext.calls) != 0 || atomic.LoadInt32(&exp.calls) != 0 {
		t.Error("disabled extractor/explainer must not be called")
	}
}

func TestSearch_SpecOverridesSkipExtraction(t *testing.T) {
	ext := &fakeExtractor{enabled: true, specs: &searchv1.ParsedSpecs{DataRate: "10G"}}
	s := newTestService(ext, &fakeEmbedder{vec: []float32{1}},
		&fakeCatalog{hits: []domain.CatalogHit{hit("a", "A", 0.1, `{"data_rate":"100G"}`)}},
		&fakeExplainer{})
	req := &searchv1.SearchRequest{Query: "q", SpecOverrides: &searchv1.ParsedSpecs{DataRate: "100G"}}
	resp, err := s.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&ext.calls) != 0 {
		t.Error("extractor must not be called when spec_overrides provided")
	}
	if resp.GetParsedSpecs().GetDataRate() != "100G" {
		t.Errorf("overrides should be echoed back, got %v", resp.GetParsedSpecs())
	}
}

func TestSpecFit(t *testing.T) {
	specs := &searchv1.ParsedSpecs{
		DataRate:      "100G",
		FormFactor:    "QSFP28",
		ReachKm:       10,
		Compatibility: []string{"Cisco Nexus"},
	}
	prod := map[string]any{
		"data_rate":     "100G",
		"form_factor":   "QSFP28",
		"reach_km":      float64(10),
		"compatibility": []any{"Cisco Nexus 9000"},
	}
	score, considered := specFit(specs, prod)
	if considered != 4 {
		t.Fatalf("considered = %d, want 4", considered)
	}
	if score != 100 {
		t.Errorf("full match score = %v, want 100", score)
	}

	// Shorter reach must fail the >= check.
	prod["reach_km"] = float64(2)
	if score, _ := specFit(specs, prod); score != 75 {
		t.Errorf("3/4 match score = %v, want 75", score)
	}
}

func TestClampLimit(t *testing.T) {
	cases := map[int32]int{0: defaultLimit, -3: defaultLimit, 5: 5, 999: maxLimit}
	for in, want := range cases {
		if got := clampLimit(in); got != want {
			t.Errorf("clampLimit(%d) = %d, want %d", in, got, want)
		}
	}
}
