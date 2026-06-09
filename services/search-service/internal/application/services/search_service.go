// Package services holds the hybrid search orchestration. It depends only on
// the interfaces declared here; concrete adapters (Anthropic, Postgres,
// embeddings, Redis) live under internal/infrastructure and satisfy them
// structurally. Pipeline: extract specs (Claude tool-use, cached) → embed query
// (cached) → pgvector candidate retrieval → spec-fit scoring/sort → match
// explanations (Claude, parallel, cached).
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
	"search-service/internal/domain"
	"search-service/internal/infrastructure/embedding"
	"search-service/pkg/logger"
)

const (
	defaultLimit        = 20
	maxLimit            = 50
	candidateMultiplier = 4  // retrieve N× the requested limit as vector candidates
	maxExplanations     = 5  // only the top-N hits get a Claude rationale
)

// SpecExtractor turns a natural-language query into structured specs.
type SpecExtractor interface {
	ExtractSpecs(ctx context.Context, query string) (*searchv1.ParsedSpecs, error)
	Enabled() bool
}

// Embedder produces query embeddings (same contract as the seed-time client).
type Embedder interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
}

// CatalogRepo retrieves nearest products by vector distance.
type CatalogRepo interface {
	VectorSearch(ctx context.Context, queryVec []float32, limit int) ([]domain.CatalogHit, error)
}

// Explainer writes a one-line rationale for a single hit.
type Explainer interface {
	Explain(ctx context.Context, query string, hit domain.CatalogHit) (string, error)
	Enabled() bool
}

// Cache is the optional JSON cache for the expensive AI steps.
type Cache interface {
	GetJSON(ctx context.Context, key string, dst any) (bool, error)
	SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error
}

// NoopCache is used when Redis isn't configured: every Get misses, every Set
// is dropped. Keeps the pipeline branch-free.
type NoopCache struct{}

func (NoopCache) GetJSON(context.Context, string, any) (bool, error)        { return false, nil }
func (NoopCache) SetJSON(context.Context, string, any, time.Duration) error { return nil }

type SearchService struct {
	extractor SpecExtractor
	embedder  Embedder
	catalog   CatalogRepo
	explainer Explainer
	cache     Cache
	cacheTTL  time.Duration
	log       *logger.Logger
}

func NewSearchService(
	extractor SpecExtractor,
	embedder Embedder,
	catalog CatalogRepo,
	explainer Explainer,
	cache Cache,
	cacheTTL time.Duration,
	log *logger.Logger,
) *SearchService {
	if cache == nil {
		cache = NoopCache{}
	}
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour
	}
	return &SearchService{
		extractor: extractor,
		embedder:  embedder,
		catalog:   catalog,
		explainer: explainer,
		cache:     cache,
		cacheTTL:  cacheTTL,
		log:       log,
	}
}

func (s *SearchService) Search(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	if query == "" {
		return &searchv1.SearchResponse{}, nil
	}
	limit := clampLimit(req.GetLimit())
	queryID := newQueryID()

	// 1. Specs: caller overrides (edited chips) take priority; otherwise extract.
	specs := req.GetSpecOverrides()
	if specs == nil {
		specs = s.extractSpecs(ctx, query)
	}

	// 2. Query embedding (essential; cached by query hash).
	vec, err := s.embedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	// 3. Vector candidate retrieval (over-fetch, then re-rank by spec fit).
	candidates := limit * candidateMultiplier
	if candidates < defaultLimit {
		candidates = defaultLimit
	}
	hits, err := s.catalog.VectorSearch(ctx, vec, candidates)
	if err != nil {
		return nil, err
	}

	// 4. Score, sort, truncate.
	scored := scoreAndSort(hits, specs)
	if len(scored) > limit {
		scored = scored[:limit]
	}

	// 5. Match explanations for the top results (best-effort, parallel).
	s.attachExplanations(ctx, query, scored)

	return &searchv1.SearchResponse{
		ParsedSpecs: specs,
		Results:     toProtoHits(scored),
		QueryId:     queryID,
	}, nil
}

// extractSpecs is best-effort: any failure (no key, API error) degrades to nil
// specs so the pipeline still returns vector results.
func (s *SearchService) extractSpecs(ctx context.Context, query string) *searchv1.ParsedSpecs {
	if s.extractor == nil || !s.extractor.Enabled() {
		return nil
	}
	key := cacheKey("spec:", query)
	var cached searchv1.ParsedSpecs
	if hit, _ := s.cache.GetJSON(ctx, key, &cached); hit {
		return &cached
	}
	specs, err := s.extractor.ExtractSpecs(ctx, query)
	if err != nil {
		s.log.Warn("spec extraction: " + err.Error())
		return nil
	}
	_ = s.cache.SetJSON(ctx, key, specs, s.cacheTTL)
	return specs
}

func (s *SearchService) embedQuery(ctx context.Context, query string) ([]float32, error) {
	key := cacheKey("emb:", query)
	var cached []float32
	if hit, _ := s.cache.GetJSON(ctx, key, &cached); hit && len(cached) > 0 {
		return cached, nil
	}
	vecs, err := s.embedder.Embed(ctx, []string{query}, embedding.InputTypeQuery)
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return nil, errEmptyEmbedding
	}
	_ = s.cache.SetJSON(ctx, key, vecs[0], s.cacheTTL)
	return vecs[0], nil
}

func (s *SearchService) attachExplanations(ctx context.Context, query string, scored []scoredHit) {
	if s.explainer == nil || !s.explainer.Enabled() || len(scored) == 0 {
		return
	}
	n := min(maxExplanations, len(scored))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := cacheKey("explain:", query+"|"+scored[i].hit.ID)
			var cached string
			if hit, _ := s.cache.GetJSON(ctx, key, &cached); hit && cached != "" {
				scored[i].explanation = cached
				return
			}
			text, err := s.explainer.Explain(ctx, query, scored[i].hit)
			if err != nil {
				s.log.Debug("explain " + scored[i].hit.SKU + ": " + err.Error())
				return
			}
			scored[i].explanation = text
			_ = s.cache.SetJSON(ctx, key, text, s.cacheTTL)
		}(i)
	}
	wg.Wait()
}

func clampLimit(l int32) int {
	n := int(l)
	if n <= 0 {
		return defaultLimit
	}
	if n > maxLimit {
		return maxLimit
	}
	return n
}

func cacheKey(prefix, input string) string {
	sum := sha256.Sum256([]byte(input))
	return prefix + hex.EncodeToString(sum[:])
}

func newQueryID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}
