package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"post-service/internal/domain/entities"
	"post-service/internal/infrastructure/postgres/sqlcgen"
)

// ErrNoRows is surfaced to the application layer so it can map missing
// rows to NotFound without importing pgx.
var ErrNoRows = errors.New("no rows")

type RFQRepository struct {
	queries *sqlcgen.Queries
}

func NewRFQRepository(pool *pgxpool.Pool) *RFQRepository {
	return &RFQRepository{queries: sqlcgen.New(pool)}
}

func (r *RFQRepository) NextRFQSeq(ctx context.Context) (int64, error) {
	return r.queries.NextRFQSeq(ctx)
}

func (r *RFQRepository) NextQuoteSeq(ctx context.Context) (int64, error) {
	return r.queries.NextQuoteSeq(ctx)
}

func (r *RFQRepository) CreateRFQ(ctx context.Context, rfq *entities.RFQ) (*entities.RFQ, error) {
	productIDs, err := toUUIDs(rfq.MatchedProductIDs)
	if err != nil {
		return nil, fmt.Errorf("invalid matched product id: %w", err)
	}

	specs := rfq.ParsedSpecs
	if len(specs) == 0 {
		specs = []byte("{}")
	}

	row, err := r.queries.CreateRFQ(ctx, sqlcgen.CreateRFQParams{
		ID:                rfq.ID,
		BuyerID:           rfq.BuyerID,
		BuyerEmail:        nullableString(rfq.BuyerEmail),
		BuyerCompany:      nullableString(rfq.BuyerCompany),
		QueryText:         rfq.QueryText,
		ParsedSpecs:       specs,
		MatchedProductIds: productIDs,
		Status:            rfq.Status,
		Qty:               nullableInt32(rfq.Qty),
		TargetDate:        toDate(rfq.TargetDate),
		ShippingAddress:   nullableString(rfq.ShippingAddress),
		Notes:             nullableString(rfq.Notes),
		ProjectID:         nullableString(rfq.ProjectID),
	})
	if err != nil {
		return nil, err
	}
	return rfqFromRow(row), nil
}

func (r *RFQRepository) GetRFQByID(ctx context.Context, id string) (*entities.RFQ, error) {
	row, err := r.queries.GetRFQByID(ctx, id)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return rfqFromRow(row), nil
}

func (r *RFQRepository) ListRFQsByBuyer(ctx context.Context, buyerID, status string, limit, offset int32) ([]*entities.RFQ, error) {
	rows, err := r.queries.ListRFQsByBuyer(ctx, sqlcgen.ListRFQsByBuyerParams{
		BuyerID: buyerID,
		Limit:   limit,
		Offset:  offset,
		Status:  nullableString(status),
	})
	if err != nil {
		return nil, err
	}
	rfqs := make([]*entities.RFQ, 0, len(rows))
	for _, row := range rows {
		rfqs = append(rfqs, rfqFromRow(row))
	}
	return rfqs, nil
}

func (r *RFQRepository) CountRFQsByBuyer(ctx context.Context, buyerID, status string) (int32, error) {
	return r.queries.CountRFQsByBuyer(ctx, sqlcgen.CountRFQsByBuyerParams{
		BuyerID: buyerID,
		Status:  nullableString(status),
	})
}

func (r *RFQRepository) ListOpenRFQs(ctx context.Context, limit, offset int32) ([]*entities.RFQ, error) {
	rows, err := r.queries.ListOpenRFQs(ctx, sqlcgen.ListOpenRFQsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	rfqs := make([]*entities.RFQ, 0, len(rows))
	for _, row := range rows {
		rfqs = append(rfqs, rfqFromRow(row))
	}
	return rfqs, nil
}

func (r *RFQRepository) CountOpenRFQs(ctx context.Context) (int32, error) {
	return r.queries.CountOpenRFQs(ctx)
}

func (r *RFQRepository) UpdateRFQStatus(ctx context.Context, id, status string) (*entities.RFQ, error) {
	row, err := r.queries.UpdateRFQStatus(ctx, sqlcgen.UpdateRFQStatusParams{ID: id, Status: status})
	if err != nil {
		return nil, mapNoRows(err)
	}
	return rfqFromRow(row), nil
}

func (r *RFQRepository) CreatePendingQuote(ctx context.Context, quote *entities.Quote) (*entities.Quote, error) {
	supplierID, err := toUUID(quote.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("invalid supplier id: %w", err)
	}
	var productID pgtype.UUID
	if quote.ProductID != "" {
		productID, err = toUUID(quote.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product id: %w", err)
		}
	}

	row, err := r.queries.CreatePendingQuote(ctx, sqlcgen.CreatePendingQuoteParams{
		ID:         quote.ID,
		RfqID:      quote.RFQID,
		SupplierID: supplierID,
		ProductID:  productID,
		MatchScore: nullableInt32(quote.MatchScore),
	})
	if err != nil {
		return nil, err
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) ListQuotesForRFQ(ctx context.Context, rfqID string) ([]*entities.Quote, error) {
	rows, err := r.queries.ListQuotesForRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	quotes := make([]*entities.Quote, 0, len(rows))
	for _, row := range rows {
		quotes = append(quotes, quoteFromRow(row))
	}
	return quotes, nil
}

func (r *RFQRepository) GetQuoteForSupplier(ctx context.Context, rfqID, supplierID string) (*entities.Quote, error) {
	sid, err := toUUID(supplierID)
	if err != nil {
		return nil, fmt.Errorf("invalid supplier id: %w", err)
	}
	row, err := r.queries.GetQuoteForSupplier(ctx, sqlcgen.GetQuoteForSupplierParams{RfqID: rfqID, SupplierID: sid})
	if err != nil {
		return nil, mapNoRows(err)
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) SubmitQuote(ctx context.Context, rfqID, supplierID string, priceUSD float64, leadTimeDays int32, validityDate, supplierNotes string) (*entities.Quote, error) {
	sid, err := toUUID(supplierID)
	if err != nil {
		return nil, fmt.Errorf("invalid supplier id: %w", err)
	}
	price, err := toNumeric(priceUSD)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}
	row, err := r.queries.SubmitQuote(ctx, sqlcgen.SubmitQuoteParams{
		RfqID:         rfqID,
		SupplierID:    sid,
		PriceUsd:      price,
		LeadTimeDays:  nullableInt32(leadTimeDays),
		ValidityDate:  toDate(validityDate),
		SupplierNotes: nullableString(supplierNotes),
	})
	if err != nil {
		return nil, mapNoRows(err)
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) InsertManufacturerQuote(ctx context.Context, quote *entities.Quote) (*entities.Quote, error) {
	var productID pgtype.UUID
	if quote.ProductID != "" {
		var err error
		if productID, err = toUUID(quote.ProductID); err != nil {
			return nil, fmt.Errorf("invalid product id: %w", err)
		}
	}
	price, err := toNumeric(quote.PriceUSD)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}
	row, err := r.queries.InsertManufacturerQuote(ctx, sqlcgen.InsertManufacturerQuoteParams{
		ID:             quote.ID,
		RfqID:          quote.RFQID,
		ManufacturerID: nullableString(quote.ManufacturerID),
		ProductID:      productID,
		PriceUsd:       price,
		LeadTimeDays:   nullableInt32(quote.LeadTimeDays),
		ValidityDate:   toDate(quote.ValidityDate),
		SupplierNotes:  nullableString(quote.SupplierNotes),
	})
	if err != nil {
		return nil, err
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) GetQuoteByID(ctx context.Context, id string) (*entities.Quote, error) {
	row, err := r.queries.GetQuoteByID(ctx, id)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) AcceptQuote(ctx context.Context, quoteID, rfqID string) (*entities.Quote, error) {
	row, err := r.queries.AcceptQuote(ctx, sqlcgen.AcceptQuoteParams{ID: quoteID, RfqID: rfqID})
	if err != nil {
		return nil, mapNoRows(err)
	}
	return quoteFromRow(row), nil
}

func (r *RFQRepository) RejectOtherQuotes(ctx context.Context, rfqID, keepQuoteID string) error {
	return r.queries.RejectOtherQuotes(ctx, sqlcgen.RejectOtherQuotesParams{RfqID: rfqID, ID: keepQuoteID})
}

func (r *RFQRepository) ListProductsByIDs(ctx context.Context, ids []string) ([]*entities.MatchedProduct, error) {
	uuids, err := toUUIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	rows, err := r.queries.ListProductsByIDs(ctx, uuids)
	if err != nil {
		return nil, err
	}
	products := make([]*entities.MatchedProduct, 0, len(rows))
	for _, row := range rows {
		products = append(products, &entities.MatchedProduct{
			ID:           uuidString(row.ID),
			SupplierID:   uuidString(row.SupplierID),
			SKU:          row.Sku,
			Name:         row.Name,
			NameZh:       deref(row.NameZh),
			Category:     row.Category,
			Specs:        row.Specs,
			PriceUSD:     numericFloat(row.PriceUsd),
			MOQ:          derefInt32(row.Moq),
			LeadTimeDays: derefInt32(row.LeadTimeDays),
		})
	}
	return products, nil
}

func (r *RFQRepository) ListSuppliersByIDs(ctx context.Context, ids []string) ([]*entities.SupplierContact, error) {
	uuids, err := toUUIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("invalid supplier id: %w", err)
	}
	rows, err := r.queries.ListSuppliersByIDs(ctx, uuids)
	if err != nil {
		return nil, err
	}
	suppliers := make([]*entities.SupplierContact, 0, len(rows))
	for _, row := range rows {
		suppliers = append(suppliers, &entities.SupplierContact{
			ID:           uuidString(row.ID),
			Name:         row.Name,
			NameZh:       deref(row.NameZh),
			City:         row.City,
			ContactEmail: deref(row.ContactEmail),
			Verified:     row.VerifiedAt.Valid,
		})
	}
	return suppliers, nil
}

func rfqFromRow(row sqlcgen.Rfq) *entities.RFQ {
	ids := make([]string, 0, len(row.MatchedProductIds))
	for _, id := range row.MatchedProductIds {
		ids = append(ids, uuidString(id))
	}
	return &entities.RFQ{
		ID:                row.ID,
		BuyerID:           row.BuyerID,
		BuyerEmail:        deref(row.BuyerEmail),
		BuyerCompany:      deref(row.BuyerCompany),
		QueryText:         row.QueryText,
		ParsedSpecs:       row.ParsedSpecs,
		MatchedProductIDs: ids,
		Status:            row.Status,
		Qty:               derefInt32(row.Qty),
		TargetDate:        dateString(row.TargetDate),
		ShippingAddress:   deref(row.ShippingAddress),
		Notes:             deref(row.Notes),
		ProjectID:         deref(row.ProjectID),
		CreatedAt:         row.CreatedAt.Time,
	}
}

func quoteFromRow(row sqlcgen.Quote) *entities.Quote {
	var submittedAt time.Time
	if row.SubmittedAt.Valid {
		submittedAt = row.SubmittedAt.Time
	}
	return &entities.Quote{
		ID:             row.ID,
		RFQID:          row.RfqID,
		SupplierID:     uuidString(row.SupplierID),
		ManufacturerID: deref(row.ManufacturerID),
		ProductID:      uuidString(row.ProductID),
		PriceUSD:       numericFloat(row.PriceUsd),
		LeadTimeDays:   derefInt32(row.LeadTimeDays),
		ValidityDate:   dateString(row.ValidityDate),
		SupplierNotes:  deref(row.SupplierNotes),
		MatchScore:     derefInt32(row.MatchScore),
		Status:         row.Status,
		SubmittedAt:    submittedAt,
		CreatedAt:      row.CreatedAt.Time,
	}
}

func mapNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoRows
	}
	return err
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableInt32(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toUUIDs(ids []string) ([]pgtype.UUID, error) {
	out := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := toUUID(id)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func uuidString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	v, err := u.Value()
	if err != nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toDate(iso string) pgtype.Date {
	if iso == "" {
		return pgtype.Date{}
	}
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: t, Valid: true}
}

func dateString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func toNumeric(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%.2f", v)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func numericFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	f := new(big.Float).SetInt(n.Int)
	if n.Exp != 0 {
		exp := new(big.Float).SetFloat64(1)
		ten := big.NewFloat(10)
		absExp := n.Exp
		if absExp < 0 {
			absExp = -absExp
		}
		for i := int32(0); i < absExp; i++ {
			exp.Mul(exp, ten)
		}
		if n.Exp > 0 {
			f.Mul(f, exp)
		} else {
			f.Quo(f, exp)
		}
	}
	out, _ := f.Float64()
	return out
}
