// cmd/embed generates vector embeddings for catalog products with embedding IS NULL.
//
// Run from the service dir:
//   DATABASE_URL=postgres://postgres:PASS@localhost:5432/postdb?sslmode=disable \
//   VOYAGE_API_KEY=... OPENAI_API_KEY=... \
//       go run ./cmd/embed
//
// Or via compose: `make embed`. See the project Makefile.
//
// Voyage AI's voyage-3 is the primary path (1024-dim, tuned for technical
// content). If VOYAGE_API_KEY is unset or every retry fails, the CLI falls
// back to OpenAI text-embedding-3-small with dimensions=1024 so both producers
// emit the same width as the Postgres vector(1024) column.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"

	"post-service/internal/infrastructure/embedding"
	"post-service/internal/infrastructure/postgres/sqlcgen"
)

const (
	batchSize          = 32
	defaultLimit       = 1000
	voyageRatePer1MTok = 0.06 // USD; rough — keep in sync with Voyage's pricing page
)

func main() {
	limit := flag.Int("limit", defaultLimit, "max products to embed in this run")
	dryRun := flag.Bool("dry-run", false, "print embedding texts without calling any API or writing to DB")
	fake := flag.Bool("fake", false, "use deterministic hash-based vectors instead of calling Voyage/OpenAI; useful for local dev when API keys aren't configured")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	q := sqlcgen.New(conn)
	rows, err := q.ListProductsMissingEmbeddingWithSupplier(ctx, int32(*limit))
	if err != nil {
		log.Fatalf("list products: %v", err)
	}
	if len(rows) == 0 {
		log.Println("no products need embedding — nothing to do")
		return
	}
	log.Printf("found %d products needing embedding", len(rows))

	// Resolve clients lazily — operator may have only one provider configured.
	var primary, fallback embedding.Client
	if *fake {
		primary = embedding.NewFakeClient()
	} else {
		if k := os.Getenv("VOYAGE_API_KEY"); k != "" && k != "replace-with-voyage-api-key" {
			primary = embedding.NewVoyageClient(k)
		}
		if k := os.Getenv("OPENAI_API_KEY"); k != "" && k != "replace-with-openai-api-key" {
			fallback = embedding.NewOpenAIClient(k)
		}
	}
	if !*dryRun && primary == nil && fallback == nil {
		log.Fatal("no embedding provider configured: set VOYAGE_API_KEY and/or OPENAI_API_KEY (or pass -fake / -dry-run)")
	}

	start := time.Now()
	embedded := 0
	approxTokens := 0

	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		texts := make([]string, len(batch))
		for j, r := range batch {
			texts[j] = embedding.BuildEmbeddingText(embedding.ProductForEmbedding{
				Name:            r.Name,
				NameZh:          derefString(r.NameZh),
				Category:        r.Category,
				Specs:           r.Specs,
				SupplierName:    r.SupplierName,
				SupplierCity:    r.SupplierCity,
				SupplierCluster: derefString(r.SupplierCluster),
			})
			approxTokens += approxTokenCount(texts[j])
		}

		if *dryRun {
			for j, r := range batch {
				log.Printf("[dry-run] %s | %s", r.Sku, texts[j])
			}
			embedded += len(batch)
			continue
		}

		vecs, used, err := embedBatch(ctx, primary, fallback, texts)
		if err != nil {
			log.Fatalf("embed batch %d-%d: %v", i, end, err)
		}
		for j, vec := range vecs {
			v := pgvector.NewVector(vec)
			if err := q.UpdateProductEmbedding(ctx, sqlcgen.UpdateProductEmbeddingParams{
				ID:        batch[j].ID,
				Embedding: &v,
			}); err != nil {
				log.Fatalf("update embedding for %s: %v", batch[j].Sku, err)
			}
		}
		embedded += len(batch)
		log.Printf("batch %d-%d/%d via %s (%.1fs)", i+1, end, len(rows), used, time.Since(start).Seconds())
	}

	dur := time.Since(start)
	estCost := float64(approxTokens) / 1_000_000 * voyageRatePer1MTok
	log.Printf("done: %d embedded, ~%d tokens, ~$%.4f at Voyage rate, %.1fs", embedded, approxTokens, estCost, dur.Seconds())
}

// embedBatch tries the primary client first, falling back on any error.
func embedBatch(ctx context.Context, primary, fallback embedding.Client, texts []string) ([][]float32, string, error) {
	if primary != nil {
		vecs, err := primary.Embed(ctx, texts, embedding.InputTypeDocument)
		if err == nil {
			return vecs, primary.Name(), nil
		}
		log.Printf("primary (%s) failed: %v — falling back", primary.Name(), err)
	}
	if fallback != nil {
		vecs, err := fallback.Embed(ctx, texts, embedding.InputTypeDocument)
		if err != nil {
			return nil, "", fmt.Errorf("fallback also failed: %w", err)
		}
		return vecs, fallback.Name(), nil
	}
	return nil, "", fmt.Errorf("no working embedding provider")
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// approxTokenCount is a coarse heuristic for cost reporting. ~4 chars per
// token is the rule of thumb for English; we round up slightly to be a bit
// conservative in the cost line.
func approxTokenCount(s string) int {
	if len(s) == 0 {
		return 0
	}
	return (len(s) + 3) / 4
}
