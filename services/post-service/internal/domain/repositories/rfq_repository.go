package repositories

import (
	"context"
	"post-service/internal/domain/entities"
)

type RFQRepository interface {
	// NextRFQSeq / NextQuoteSeq feed the NNNN segment of formatted ids.
	NextRFQSeq(ctx context.Context) (int64, error)
	NextQuoteSeq(ctx context.Context) (int64, error)

	CreateRFQ(ctx context.Context, rfq *entities.RFQ) (*entities.RFQ, error)
	GetRFQByID(ctx context.Context, id string) (*entities.RFQ, error)
	ListRFQsByBuyer(ctx context.Context, buyerID, status string, limit, offset int32) ([]*entities.RFQ, error)
	CountRFQsByBuyer(ctx context.Context, buyerID, status string) (int32, error)
	ListOpenRFQs(ctx context.Context, limit, offset int32) ([]*entities.RFQ, error)
	CountOpenRFQs(ctx context.Context) (int32, error)
	UpdateRFQStatus(ctx context.Context, id, status string) (*entities.RFQ, error)

	CreatePendingQuote(ctx context.Context, quote *entities.Quote) (*entities.Quote, error)
	ListQuotesForRFQ(ctx context.Context, rfqID string) ([]*entities.Quote, error)
	GetQuoteForSupplier(ctx context.Context, rfqID, supplierID string) (*entities.Quote, error)
	SubmitQuote(ctx context.Context, rfqID, supplierID string, priceUSD float64, leadTimeDays int32, validityDate, supplierNotes string) (*entities.Quote, error)
	InsertManufacturerQuote(ctx context.Context, quote *entities.Quote) (*entities.Quote, error)
	GetQuoteByID(ctx context.Context, id string) (*entities.Quote, error)
	AcceptQuote(ctx context.Context, quoteID, rfqID string) (*entities.Quote, error)
	RejectOtherQuotes(ctx context.Context, rfqID, keepQuoteID string) error

	ListProductsByIDs(ctx context.Context, ids []string) ([]*entities.MatchedProduct, error)
	ListSuppliersByIDs(ctx context.Context, ids []string) ([]*entities.SupplierContact, error)
}
