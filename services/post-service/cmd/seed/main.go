// cmd/seed seeds the post-service Postgres with real Chinese photonics
// suppliers + transceiver SKUs. Idempotent on (suppliers.id) and
// (products.supplier_id, products.sku) — re-running upserts in place.
//
// Run from the service dir:
//   DATABASE_URL=postgres://postgres:PASS@localhost:5432/postdb?sslmode=disable \
//       go run ./cmd/seed -seeds ./seeds
//
// Or via compose: `make seed`. See the project Makefile.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/yaml.v3"

	"post-service/internal/infrastructure/postgres/sqlcgen"
)

// fiberlaneSupplierNamespace is the fixed UUIDv5 namespace for supplier slugs.
// Generated once and committed; do NOT rotate, or all existing supplier rows
// will appear as new on the next seed run.
var fiberlaneSupplierNamespace = uuid.MustParse("8b8e7c41-1c5f-4b1a-9e1e-3e6c4d7a0f01")

// ----- YAML schemas -----

type supplierSeed struct {
	Slug           string   `yaml:"slug"`
	Name           string   `yaml:"name"`
	NameZh         string   `yaml:"name_zh,omitempty"`
	City           string   `yaml:"city"`
	Cluster        string   `yaml:"cluster,omitempty"`
	Country        string   `yaml:"country,omitempty"`
	Capabilities   []string `yaml:"capabilities,omitempty"`
	Certifications []string `yaml:"certifications,omitempty"`
	FoundedYear    int32    `yaml:"founded_year,omitempty"`
	Employees      int32    `yaml:"employees,omitempty"`
	FacilitySizeM2 int32    `yaml:"facility_size_m2,omitempty"`
	AnnualOutput   string   `yaml:"annual_output,omitempty"`
	OnTimeRate     float64  `yaml:"on_time_rate,omitempty"`
	Rating         float64  `yaml:"rating,omitempty"`
	VerifiedAt     string   `yaml:"verified_at,omitempty"` // RFC3339
	AuditReportURL string   `yaml:"audit_report_url,omitempty"`
	PhotoURL       string   `yaml:"photo_url,omitempty"`
	ContactEmail   string   `yaml:"contact_email,omitempty"`
	ContactWechat  string   `yaml:"contact_wechat,omitempty"`
}

type productSeed struct {
	SKU          string         `yaml:"sku"`
	Name         string         `yaml:"name"`
	NameZh       string         `yaml:"name_zh,omitempty"`
	Category     string         `yaml:"category"`
	Specs        map[string]any `yaml:"specs"`
	PriceUsd     float64        `yaml:"price_usd,omitempty"`
	Moq          int32          `yaml:"moq,omitempty"`
	StockQty     int32          `yaml:"stock_qty,omitempty"`
	LeadTimeDays int32          `yaml:"lead_time_days,omitempty"`
	DatasheetURL string         `yaml:"datasheet_url,omitempty"`
}

type productSeedFile struct {
	SupplierSlug string        `yaml:"supplier_slug"`
	Products     []productSeed `yaml:"products"`
}

// ----- helpers -----

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilIfZero(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

// numericFromFloat returns a pgtype.Numeric NULL when v is 0, otherwise the
// value scaled to two decimals (enough for prices, on-time rates, ratings).
func numericFromFloat(v float64) pgtype.Numeric {
	if v == 0 {
		return pgtype.Numeric{}
	}
	// Convert via pgx-friendly string form. pgtype.Numeric.Scan accepts strings
	// with arbitrary precision; this avoids float-to-decimal drift better than
	// constructing from big.Int directly.
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.4f", v))
	return n
}

// Suppress unused import lint when big is not used in the future.
var _ = big.NewInt

func tsFromRFC3339(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Fatalf("invalid timestamp %q: %v", s, err)
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgUUID(u uuid.UUID) pgtype.UUID {
	var out pgtype.UUID
	copy(out.Bytes[:], u[:])
	out.Valid = true
	return out
}

// supplierUUID derives a deterministic UUIDv5 from the supplier slug. Same slug
// → same UUID across re-runs, which is what lets UpsertSupplier be idempotent
// without adding a UNIQUE column on suppliers.name.
func supplierUUID(slug string) uuid.UUID {
	return uuid.NewSHA1(fiberlaneSupplierNamespace, []byte(strings.ToLower(slug)))
}

// ----- main -----

func main() {
	seedsDir := flag.String("seeds", "./seeds", "path to the seeds directory")
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

	supplierIDs, err := seedSuppliers(ctx, q, filepath.Join(*seedsDir, "suppliers"))
	if err != nil {
		log.Fatalf("seed suppliers: %v", err)
	}
	productCount, err := seedProducts(ctx, q, filepath.Join(*seedsDir, "products"), supplierIDs)
	if err != nil {
		log.Fatalf("seed products: %v", err)
	}

	log.Printf("seed complete: %d suppliers, %d products", len(supplierIDs), productCount)
}

func seedSuppliers(ctx context.Context, q *sqlcgen.Queries, dir string) (map[string]uuid.UUID, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	out := map[string]uuid.UUID{}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var s supplierSeed
		if err := yaml.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if s.Slug == "" || s.Name == "" || s.City == "" {
			return nil, fmt.Errorf("%s: slug, name, city are required", path)
		}
		country := s.Country
		if country == "" {
			country = "CN"
		}
		id := supplierUUID(s.Slug)

		_, err = q.UpsertSupplier(ctx, sqlcgen.UpsertSupplierParams{
			ID:             pgUUID(id),
			Name:           s.Name,
			NameZh:         nilIfEmpty(s.NameZh),
			City:           s.City,
			Country:        country,
			Cluster:        nilIfEmpty(s.Cluster),
			Capabilities:   s.Capabilities,
			Certifications: s.Certifications,
			FoundedYear:    nilIfZero(s.FoundedYear),
			Employees:      nilIfZero(s.Employees),
			FacilitySizeM2: nilIfZero(s.FacilitySizeM2),
			AnnualOutput:   nilIfEmpty(s.AnnualOutput),
			OnTimeRate:     numericFromFloat(s.OnTimeRate),
			Rating:         numericFromFloat(s.Rating),
			VerifiedAt:     tsFromRFC3339(s.VerifiedAt),
			AuditReportUrl: nilIfEmpty(s.AuditReportURL),
			PhotoUrl:       nilIfEmpty(s.PhotoURL),
			ContactEmail:   nilIfEmpty(s.ContactEmail),
			ContactWechat: nilIfEmpty(s.ContactWechat),
		})
		if err != nil {
			return nil, fmt.Errorf("upsert supplier %s: %w", s.Slug, err)
		}
		out[s.Slug] = id
		log.Printf("supplier  %-22s %s", s.Slug, id)
	}
	return out, nil
}

func seedProducts(ctx context.Context, q *sqlcgen.Queries, dir string, suppliers map[string]uuid.UUID) (int, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return 0, err
	}
	count := 0
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", path, err)
		}
		var f productSeedFile
		if err := yaml.Unmarshal(raw, &f); err != nil {
			return count, fmt.Errorf("parse %s: %w", path, err)
		}
		supplierID, ok := suppliers[f.SupplierSlug]
		if !ok {
			return count, fmt.Errorf("%s: unknown supplier_slug %q", path, f.SupplierSlug)
		}

		for _, p := range f.Products {
			if p.SKU == "" || p.Name == "" || p.Category == "" {
				return count, fmt.Errorf("%s: product %q missing sku/name/category", path, p.SKU)
			}
			specs, err := json.Marshal(p.Specs)
			if err != nil {
				return count, fmt.Errorf("%s: marshal specs for %s: %w", path, p.SKU, err)
			}
			_, err = q.UpsertProduct(ctx, sqlcgen.UpsertProductParams{
				SupplierID:   pgUUID(supplierID),
				Sku:          p.SKU,
				Name:         p.Name,
				NameZh:       nilIfEmpty(p.NameZh),
				Category:     p.Category,
				Specs:        specs,
				PriceUsd:     numericFromFloat(p.PriceUsd),
				Moq:          nilIfZero(p.Moq),
				StockQty:     nilIfZero(p.StockQty),
				LeadTimeDays: nilIfZero(p.LeadTimeDays),
				DatasheetUrl: nilIfEmpty(p.DatasheetURL),
			})
			if err != nil {
				return count, fmt.Errorf("upsert product %s: %w", p.SKU, err)
			}
			count++
		}
		log.Printf("products  %-22s %d sku(s)", f.SupplierSlug, len(f.Products))
	}
	return count, nil
}
