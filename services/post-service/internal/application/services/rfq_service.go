package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"post-service/internal/application/dto"
	appErrors "post-service/internal/application/errors"
	"post-service/internal/domain/entities"
	"post-service/internal/domain/repositories"
	"post-service/internal/infrastructure/messaging"
	"post-service/internal/infrastructure/postgres"
	"post-service/pkg/logger"
)

// rfqIDSuffix tags ids with the sourcing corridor (Shenzhen). Single-corridor
// MVP, so it's a constant rather than derived from the supplier cluster.
const rfqIDSuffix = "SZX"

// MagicLinkIssuer is implemented by the auth-service gRPC client; an
// interface keeps the usecase testable without a live auth-service.
type MagicLinkIssuer interface {
	IssueMagicLinkToken(ctx context.Context, rfqID, supplierID string) (string, error)
}

// RFQEventPublisher abstracts the RabbitMQ publisher (nil-able in dev).
type RFQEventPublisher interface {
	PublishRFQCreated(event messaging.RFQCreatedEvent) error
	PublishQuoteSubmitted(event messaging.QuoteSubmittedEvent) error
}

type RFQService struct {
	repo            repositories.RFQRepository
	magicLinks      MagicLinkIssuer
	eventPublisher  RFQEventPublisher
	frontendBaseURL string
	logger          *logger.Logger
	now             func() time.Time
}

func NewRFQService(
	repo repositories.RFQRepository,
	magicLinks MagicLinkIssuer,
	eventPublisher RFQEventPublisher,
	frontendBaseURL string,
	logger *logger.Logger,
) *RFQService {
	return &RFQService{
		repo:            repo,
		magicLinks:      magicLinks,
		eventPublisher:  eventPublisher,
		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
		logger:          logger,
		now:             time.Now,
	}
}

// CreateRFQ persists the RFQ, creates one pending quote per distinct supplier
// of the matched products, mints magic-link tokens, and publishes rfq.created
// so notification-service can email suppliers and the buyer.
func (s *RFQService) CreateRFQ(ctx context.Context, req *dto.CreateRFQRequest) (*entities.RFQ, error) {
	if req.BuyerID == "" || strings.TrimSpace(req.QueryText) == "" {
		return nil, appErrors.ErrInvalidRFQData
	}
	if len(req.MatchedProductIDs) == 0 {
		return nil, appErrors.ErrNoMatchedProducts
	}

	products, err := s.repo.ListProductsByIDs(ctx, req.MatchedProductIDs)
	if err != nil {
		s.logger.Error("rfq: resolve matched products: " + err.Error())
		return nil, appErrors.ErrInvalidRFQData
	}
	if len(products) == 0 {
		return nil, appErrors.ErrNoMatchedProducts
	}

	// One invitation per supplier; the first matched product of each supplier
	// anchors the pending quote and the part description in the email.
	productBySupplier := make(map[string]*entities.MatchedProduct)
	supplierOrder := make([]string, 0, len(products))
	for _, p := range products {
		if _, seen := productBySupplier[p.SupplierID]; !seen {
			productBySupplier[p.SupplierID] = p
			supplierOrder = append(supplierOrder, p.SupplierID)
		}
	}

	suppliers, err := s.repo.ListSuppliersByIDs(ctx, supplierOrder)
	if err != nil {
		s.logger.Error("rfq: resolve suppliers: " + err.Error())
		return nil, appErrors.ErrRFQCreationFailed
	}
	supplierByID := make(map[string]*entities.SupplierContact, len(suppliers))
	for _, sup := range suppliers {
		supplierByID[sup.ID] = sup
	}

	rfqID, err := s.nextID(ctx, "RFQ", s.repo.NextRFQSeq)
	if err != nil {
		s.logger.Error("rfq: allocate id: " + err.Error())
		return nil, appErrors.ErrRFQCreationFailed
	}

	rfq, err := s.repo.CreateRFQ(ctx, &entities.RFQ{
		ID:                rfqID,
		BuyerID:           req.BuyerID,
		BuyerEmail:        req.BuyerEmail,
		BuyerCompany:      req.BuyerCompany,
		QueryText:         strings.TrimSpace(req.QueryText),
		ParsedSpecs:       req.ParsedSpecs,
		MatchedProductIDs: req.MatchedProductIDs,
		Status:            entities.RFQStatusOpen,
		Qty:               req.Qty,
		TargetDate:        req.TargetDate,
		ShippingAddress:   req.ShippingAddress,
		Notes:             req.Notes,
	})
	if err != nil {
		s.logger.Error("rfq: insert: " + err.Error())
		return nil, appErrors.ErrRFQCreationFailed
	}

	invites := make([]messaging.RFQSupplierInvite, 0, len(supplierOrder))
	for _, supplierID := range supplierOrder {
		product := productBySupplier[supplierID]

		quoteID, seqErr := s.nextID(ctx, "QUOTE", s.repo.NextQuoteSeq)
		if seqErr != nil {
			s.logger.Error("rfq: allocate quote id: " + seqErr.Error())
			continue
		}
		if _, qErr := s.repo.CreatePendingQuote(ctx, &entities.Quote{
			ID:         quoteID,
			RFQID:      rfq.ID,
			SupplierID: supplierID,
			ProductID:  product.ID,
		}); qErr != nil {
			s.logger.Error("rfq: create pending quote for supplier " + supplierID + ": " + qErr.Error())
			continue
		}

		invite := messaging.RFQSupplierInvite{
			SupplierID:  supplierID,
			ProductSKU:  product.SKU,
			ProductName: product.Name,
		}
		if contact, ok := supplierByID[supplierID]; ok {
			invite.Name = contact.Name
			invite.NameZh = contact.NameZh
			invite.City = contact.City
			invite.ContactEmail = contact.ContactEmail
		}

		// A failed mint must not fail RFQ creation: notification-service
		// skips invites without a link, and the supplier can be re-invited.
		if s.magicLinks != nil {
			token, mlErr := s.magicLinks.IssueMagicLinkToken(ctx, rfq.ID, supplierID)
			if mlErr != nil {
				s.logger.Error("rfq: mint magic link for supplier " + supplierID + ": " + mlErr.Error())
			} else if s.frontendBaseURL != "" {
				invite.MagicLinkURL = s.frontendBaseURL + "/q/" + token
			}
		}

		invites = append(invites, invite)
	}

	s.publishRFQCreated(rfq, productBySupplier, invites)
	return rfq, nil
}

func (s *RFQService) GetRFQ(ctx context.Context, id, requestingUserID string) (*entities.RFQ, error) {
	rfq, err := s.repo.GetRFQByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrRFQNotFound
		}
		s.logger.Error("rfq: get: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if rfq.BuyerID != requestingUserID {
		return nil, appErrors.ErrUnauthorizedAccess
	}
	return rfq, nil
}

func (s *RFQService) ListRFQs(ctx context.Context, req *dto.ListRFQsRequest) ([]*entities.RFQ, int32, error) {
	if req.BuyerID == "" {
		return nil, 0, appErrors.ErrInvalidRequest
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	rfqs, err := s.repo.ListRFQsByBuyer(ctx, req.BuyerID, req.Status, limit, offset)
	if err != nil {
		s.logger.Error("rfq: list: " + err.Error())
		return nil, 0, appErrors.ErrServiceUnavailable
	}
	total, err := s.repo.CountRFQsByBuyer(ctx, req.BuyerID, req.Status)
	if err != nil {
		s.logger.Error("rfq: count: " + err.Error())
		return nil, 0, appErrors.ErrServiceUnavailable
	}
	return rfqs, total, nil
}

// ListQuotesForRFQ enforces buyer ownership before exposing quotes.
func (s *RFQService) ListQuotesForRFQ(ctx context.Context, rfqID, requestingUserID string) ([]*entities.Quote, error) {
	if _, err := s.GetRFQ(ctx, rfqID, requestingUserID); err != nil {
		return nil, err
	}
	quotes, err := s.repo.ListQuotesForRFQ(ctx, rfqID)
	if err != nil {
		s.logger.Error("rfq: list quotes: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return quotes, nil
}

// GetRFQForSupplier serves the magic-link page. The (rfqID, supplierID) pair
// comes from a token the gateway already validated with auth-service; the
// pending-quote row is the service-side proof the supplier was invited.
func (s *RFQService) GetRFQForSupplier(ctx context.Context, rfqID, supplierID string) (*entities.RFQ, *entities.Quote, string, error) {
	rfq, err := s.repo.GetRFQByID(ctx, rfqID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, nil, "", appErrors.ErrRFQNotFound
		}
		s.logger.Error("rfq: supplier view get: " + err.Error())
		return nil, nil, "", appErrors.ErrServiceUnavailable
	}

	quote, err := s.repo.GetQuoteForSupplier(ctx, rfqID, supplierID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, nil, "", appErrors.ErrQuoteNotFound
		}
		s.logger.Error("rfq: supplier view quote: " + err.Error())
		return nil, nil, "", appErrors.ErrServiceUnavailable
	}

	supplierName := ""
	if contacts, cErr := s.repo.ListSuppliersByIDs(ctx, []string{supplierID}); cErr == nil && len(contacts) > 0 {
		supplierName = contacts[0].Name
	}

	// Buyer email never reaches the supplier-facing surface.
	rfq.BuyerEmail = ""
	return rfq, quote, supplierName, nil
}

// SubmitQuote fills in the pending quote, marks the RFQ quoted, and publishes
// quote.submitted so the buyer gets notified.
func (s *RFQService) SubmitQuote(ctx context.Context, req *dto.SubmitQuoteRequest) (*entities.Quote, error) {
	if req.RFQID == "" || req.SupplierID == "" || req.PriceUSD <= 0 || req.LeadTimeDays <= 0 {
		return nil, appErrors.ErrInvalidRequest
	}

	quote, err := s.repo.SubmitQuote(ctx, req.RFQID, req.SupplierID, req.PriceUSD, req.LeadTimeDays, req.ValidityDate, req.SupplierNotes)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			// No pending row: either never invited (NotFound) or already submitted.
			if existing, exErr := s.repo.GetQuoteForSupplier(ctx, req.RFQID, req.SupplierID); exErr == nil && existing.Status != entities.QuoteStatusPending {
				return nil, appErrors.ErrQuoteAlreadyExists
			}
			return nil, appErrors.ErrQuoteNotFound
		}
		s.logger.Error("rfq: submit quote: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}

	rfq, err := s.repo.UpdateRFQStatus(ctx, req.RFQID, entities.RFQStatusQuoted)
	if err != nil {
		s.logger.Error("rfq: mark quoted: " + err.Error())
		rfq = nil
	}

	if s.eventPublisher != nil {
		event := messaging.QuoteSubmittedEvent{
			QuoteID:      quote.ID,
			RFQID:        quote.RFQID,
			SupplierID:   quote.SupplierID,
			PriceUSD:     quote.PriceUSD,
			LeadTimeDays: quote.LeadTimeDays,
			SubmittedAt:  quote.SubmittedAt,
		}
		if rfq != nil {
			event.BuyerID = rfq.BuyerID
			event.BuyerEmail = rfq.BuyerEmail
			event.BuyerCompany = rfq.BuyerCompany
			event.QueryText = rfq.QueryText
		}
		if contacts, cErr := s.repo.ListSuppliersByIDs(ctx, []string{quote.SupplierID}); cErr == nil && len(contacts) > 0 {
			event.SupplierName = contacts[0].Name
		}
		if pubErr := s.eventPublisher.PublishQuoteSubmitted(event); pubErr != nil {
			s.logger.Error("rfq: publish quote.submitted: " + pubErr.Error())
		}
	}

	return quote, nil
}

func (s *RFQService) publishRFQCreated(rfq *entities.RFQ, productBySupplier map[string]*entities.MatchedProduct, invites []messaging.RFQSupplierInvite) {
	if s.eventPublisher == nil {
		return
	}

	partSummary := rfq.QueryText
	for _, p := range productBySupplier {
		partSummary = fmt.Sprintf("%s (%s)", p.Name, p.SKU)
		break
	}

	event := messaging.RFQCreatedEvent{
		RFQID:           rfq.ID,
		BuyerID:         rfq.BuyerID,
		BuyerEmail:      rfq.BuyerEmail,
		BuyerCompany:    rfq.BuyerCompany,
		QueryText:       rfq.QueryText,
		PartSummary:     partSummary,
		Qty:             rfq.Qty,
		TargetDate:      rfq.TargetDate,
		ShippingAddress: rfq.ShippingAddress,
		Notes:           rfq.Notes,
		Suppliers:       invites,
		CreatedAt:       rfq.CreatedAt,
	}
	if err := s.eventPublisher.PublishRFQCreated(event); err != nil {
		s.logger.Error("rfq: publish rfq.created: " + err.Error())
	}
}

// nextID formats RFQ-YYYYMMDD-NNNN-SZX / QUOTE-YYYYMMDD-NNNN-SZX. NNNN comes
// from a global sequence, so ids stay unique without per-day coordination
// (it grows past 4 digits rather than wrapping).
func (s *RFQService) nextID(ctx context.Context, prefix string, next func(context.Context) (int64, error)) (string, error) {
	seq, err := next(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%04d-%s", prefix, s.now().UTC().Format("20060102"), seq, rfqIDSuffix), nil
}
