package services

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
	"google.golang.org/protobuf/types/known/structpb"

	"search-service/internal/domain"
)

var errEmptyEmbedding = errors.New("search: embedder returned empty vector")

// Ranking blend: vector closeness dominates, spec fit nudges. Tuned for the
// transceiver demo where vector recall is strong and specs disambiguate ties.
const (
	weightVector = 0.65
	weightSpec   = 0.35
)

type scoredHit struct {
	hit         domain.CatalogHit
	score       float64 // 0..100
	explanation string
}

// scoreAndSort blends vector closeness with spec fit, then sorts descending.
// When no specs were extracted, score is pure vector closeness (no penalty).
func scoreAndSort(hits []domain.CatalogHit, specs *searchv1.ParsedSpecs) []scoredHit {
	scored := make([]scoredHit, 0, len(hits))
	for _, h := range hits {
		vecScore := clamp01(1-h.VectorDistance) * 100
		prodSpecs := parseSpecs(h.SpecsJSON)
		specScore, considered := specFit(specs, prodSpecs)

		final := vecScore
		if considered > 0 {
			final = weightVector*vecScore + weightSpec*specScore
		}
		scored = append(scored, scoredHit{hit: h, score: final})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].hit.VectorDistance < scored[j].hit.VectorDistance
	})
	return scored
}

// specFit returns the fraction (0..100) of extracted specs the product
// satisfies and how many specs were considered. Considered==0 means the query
// carried no structured specs, so the caller should fall back to vector score.
func specFit(specs *searchv1.ParsedSpecs, prod map[string]any) (float64, int) {
	if specs == nil || prod == nil {
		return 0, 0
	}
	considered, matched := 0, 0

	if specs.GetDataRate() != "" {
		considered++
		if strings.EqualFold(prodStr(prod, "data_rate"), specs.GetDataRate()) {
			matched++
		}
	}
	if specs.GetFormFactor() != "" {
		considered++
		if strings.EqualFold(prodStr(prod, "form_factor"), specs.GetFormFactor()) {
			matched++
		}
	}
	if specs.GetReachKm() > 0 {
		considered++
		if prodFloat(prod, "reach_km") >= specs.GetReachKm() {
			matched++
		}
	}
	if specs.GetWavelengthNm() > 0 {
		considered++
		if int32(prodFloat(prod, "wavelength_nm")) == specs.GetWavelengthNm() {
			matched++
		}
	}
	if specs.GetFiberType() != "" {
		considered++
		if strings.EqualFold(prodStr(prod, "fiber_type"), specs.GetFiberType()) {
			matched++
		}
	}
	if len(specs.GetCompatibility()) > 0 {
		considered++
		if compatOverlap(prod["compatibility"], specs.GetCompatibility()) {
			matched++
		}
	}

	if considered == 0 {
		return 0, 0
	}
	return float64(matched) / float64(considered) * 100, considered
}

func toProtoHits(scored []scoredHit) []*searchv1.ProductHit {
	if len(scored) == 0 {
		return nil
	}
	out := make([]*searchv1.ProductHit, 0, len(scored))
	for _, sh := range scored {
		h := sh.hit
		out = append(out, &searchv1.ProductHit{
			Id:               h.ID,
			SupplierId:       h.SupplierID,
			Sku:              h.SKU,
			Name:             h.Name,
			NameZh:           h.NameZh,
			Category:         h.Category,
			Specs:            specsStruct(h.SpecsJSON),
			PriceUsd:         h.PriceUSD,
			Moq:              h.MOQ,
			StockQty:         h.StockQty,
			LeadTimeDays:     h.LeadTimeDays,
			DatasheetUrl:     h.DatasheetURL,
			SupplierName:     h.SupplierName,
			SupplierCity:     h.SupplierCity,
			SupplierVerified: h.SupplierVerified,
			MatchScore:       math.Round(sh.score),
			VectorDistance:   h.VectorDistance,
			MatchExplanation: sh.explanation,
		})
	}
	return out
}

func parseSpecs(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func specsStruct(raw []byte) *structpb.Struct {
	m := parseSpecs(raw)
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

func prodStr(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func prodFloat(m map[string]any, k string) float64 {
	if v, ok := m[k].(float64); ok {
		return v
	}
	return 0
}

// compatOverlap reports whether any wanted compatibility token appears in the
// product's compatibility list, matched loosely (substring, case-insensitive)
// so "Cisco Nexus" matches a product tagged "Cisco Nexus 9000".
func compatOverlap(raw any, want []string) bool {
	list, ok := raw.([]any)
	if !ok {
		return false
	}
	for _, item := range list {
		have, ok := item.(string)
		if !ok {
			continue
		}
		haveL := strings.ToLower(have)
		for _, w := range want {
			wl := strings.ToLower(w)
			if wl == "" {
				continue
			}
			if strings.Contains(haveL, wl) || strings.Contains(wl, haveL) {
				return true
			}
		}
	}
	return false
}

func clamp01(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}
