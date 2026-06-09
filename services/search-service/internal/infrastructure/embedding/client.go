// Package embedding wraps the external embedding-generation services used by
// Fiberlane. The interface is intentionally small (one method) so the seed CLI
// and the search service can both swap clients without code changes.
//
// Voyage AI's voyage-3 is the primary model (1024-dim, tuned for technical
// content). OpenAI's text-embedding-3-small with dimensions=1024 is the
// fallback so both produce vectors that fit our vector(1024) Postgres column.
package embedding

import "context"

// Vector dimension shared by all clients. Must match the column declared in
// the migration (services/post-service/migrations/004_products.up.sql).
const Dim = 1024

// InputTypeDocument tags catalog content (products) at index time. Voyage uses
// this hint to bias the embedding; OpenAI ignores it. The search service uses
// InputTypeQuery at retrieval time so the two-tower contrast can fire.
const (
	InputTypeDocument = "document"
	InputTypeQuery    = "query"
)

// Client embeds a batch of texts in one call. Implementations are expected to
// honor ctx for cancellation and to return an error if the response shape
// drifts from the agreed 1024-dim contract.
type Client interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
	// Name is used in logs / cost reports so the operator can tell which
	// provider actually produced a given batch.
	Name() string
}
