// Package domain holds the search service's entities and the interfaces its
// orchestration depends on. Concrete adapters (Postgres, Anthropic, Redis,
// embeddings) live under internal/infrastructure and satisfy these structurally.
package domain

// CatalogHit is one product row returned by the vector search, joined with the
// display fields the results screen needs. SpecsJSON is the raw `specs` jsonb
// kept as bytes so the application layer decides how to decode it (into a
// structpb.Struct for the wire, and into a map for spec-fit scoring).
type CatalogHit struct {
	ID               string
	SupplierID       string
	SKU              string
	Name             string
	NameZh           string
	Category         string
	SpecsJSON        []byte
	PriceUSD         float64
	MOQ              int32
	StockQty         int32
	LeadTimeDays     int32
	DatasheetURL     string
	SupplierName     string
	SupplierCity     string
	SupplierVerified bool
	// VectorDistance is the raw pgvector cosine distance (lower = closer).
	VectorDistance float64
}
