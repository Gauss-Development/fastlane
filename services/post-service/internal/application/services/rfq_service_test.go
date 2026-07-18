package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"post-service/internal/application/dto"
	appErrors "post-service/internal/application/errors"
	"post-service/internal/domain/entities"
	"post-service/internal/infrastructure/messaging"
	"post-service/internal/infrastructure/postgres"
	"post-service/pkg/logger"
)

type fakeRFQRepo struct {
	rfqSeq     int64
	quoteSeq   int64
	rfqs       map[string]*entities.RFQ
	quotes     map[string]*entities.Quote // keyed by rfqID+"|"+supplierID for supplier quotes
	quotesByID map[string]*entities.Quote // keyed by quoteID for all quotes
	products   []*entities.MatchedProduct
	suppliers  []*entities.SupplierContact
}

func newFakeRFQRepo() *fakeRFQRepo {
	return &fakeRFQRepo{
		rfqs:       map[string]*entities.RFQ{},
		quotes:     map[string]*entities.Quote{},
		quotesByID: map[string]*entities.Quote{},
	}
}

func (f *fakeRFQRepo) NextRFQSeq(context.Context) (int64, error)   { f.rfqSeq++; return f.rfqSeq, nil }
func (f *fakeRFQRepo) NextQuoteSeq(context.Context) (int64, error) { f.quoteSeq++; return f.quoteSeq, nil }

func (f *fakeRFQRepo) CreateRFQ(_ context.Context, rfq *entities.RFQ) (*entities.RFQ, error) {
	stored := *rfq
	stored.CreatedAt = time.Now()
	f.rfqs[rfq.ID] = &stored
	return &stored, nil
}

func (f *fakeRFQRepo) GetRFQByID(_ context.Context, id string) (*entities.RFQ, error) {
	rfq, ok := f.rfqs[id]
	if !ok {
		return nil, postgres.ErrNoRows
	}
	return rfq, nil
}

func (f *fakeRFQRepo) ListRFQsByBuyer(_ context.Context, buyerID, status string, _, _ int32) ([]*entities.RFQ, error) {
	out := []*entities.RFQ{}
	for _, rfq := range f.rfqs {
		if rfq.BuyerID == buyerID && (status == "" || rfq.Status == status) {
			out = append(out, rfq)
		}
	}
	return out, nil
}

func (f *fakeRFQRepo) CountRFQsByBuyer(ctx context.Context, buyerID, status string) (int32, error) {
	rfqs, _ := f.ListRFQsByBuyer(ctx, buyerID, status, 0, 0)
	return int32(len(rfqs)), nil
}

func (f *fakeRFQRepo) ListOpenRFQs(_ context.Context, _, _ int32) ([]*entities.RFQ, error) {
	out := []*entities.RFQ{}
	for _, rfq := range f.rfqs {
		if rfq.Status == entities.RFQStatusOpen {
			out = append(out, rfq)
		}
	}
	return out, nil
}

func (f *fakeRFQRepo) CountOpenRFQs(ctx context.Context) (int32, error) {
	rfqs, _ := f.ListOpenRFQs(ctx, 0, 0)
	return int32(len(rfqs)), nil
}

func (f *fakeRFQRepo) UpdateRFQStatus(_ context.Context, id, status string) (*entities.RFQ, error) {
	rfq, ok := f.rfqs[id]
	if !ok {
		return nil, postgres.ErrNoRows
	}
	rfq.Status = status
	return rfq, nil
}

func (f *fakeRFQRepo) CreatePendingQuote(_ context.Context, quote *entities.Quote) (*entities.Quote, error) {
	stored := *quote
	stored.Status = entities.QuoteStatusPending
	stored.CreatedAt = time.Now()
	f.quotes[quote.RFQID+"|"+quote.SupplierID] = &stored
	f.quotesByID[quote.ID] = &stored
	return &stored, nil
}

func (f *fakeRFQRepo) ListQuotesForRFQ(_ context.Context, rfqID string) ([]*entities.Quote, error) {
	out := []*entities.Quote{}
	for _, quote := range f.quotes {
		if quote.RFQID == rfqID {
			out = append(out, quote)
		}
	}
	return out, nil
}

func (f *fakeRFQRepo) GetQuoteForSupplier(_ context.Context, rfqID, supplierID string) (*entities.Quote, error) {
	quote, ok := f.quotes[rfqID+"|"+supplierID]
	if !ok {
		return nil, postgres.ErrNoRows
	}
	return quote, nil
}

func (f *fakeRFQRepo) SubmitQuote(_ context.Context, rfqID, supplierID string, priceUSD float64, leadTimeDays int32, validityDate, supplierNotes string) (*entities.Quote, error) {
	quote, ok := f.quotes[rfqID+"|"+supplierID]
	if !ok || quote.Status != entities.QuoteStatusPending {
		return nil, postgres.ErrNoRows
	}
	quote.PriceUSD = priceUSD
	quote.LeadTimeDays = leadTimeDays
	quote.ValidityDate = validityDate
	quote.SupplierNotes = supplierNotes
	quote.Status = entities.QuoteStatusSubmitted
	quote.SubmittedAt = time.Now()
	f.quotesByID[quote.ID] = quote
	return quote, nil
}

func (f *fakeRFQRepo) InsertManufacturerQuote(_ context.Context, quote *entities.Quote) (*entities.Quote, error) {
	stored := *quote
	stored.Status = entities.QuoteStatusSubmitted
	stored.SubmittedAt = time.Now()
	stored.CreatedAt = time.Now()
	f.quotesByID[quote.ID] = &stored
	return &stored, nil
}

func (f *fakeRFQRepo) GetQuoteByID(_ context.Context, id string) (*entities.Quote, error) {
	q, ok := f.quotesByID[id]
	if !ok {
		return nil, postgres.ErrNoRows
	}
	return q, nil
}

func (f *fakeRFQRepo) AcceptQuote(_ context.Context, quoteID, rfqID string) (*entities.Quote, error) {
	q, ok := f.quotesByID[quoteID]
	if !ok || q.RFQID != rfqID || q.Status != entities.QuoteStatusSubmitted {
		return nil, postgres.ErrNoRows
	}
	q.Status = entities.QuoteStatusAccepted
	return q, nil
}

func (f *fakeRFQRepo) RejectOtherQuotes(_ context.Context, rfqID, keepQuoteID string) error {
	for _, q := range f.quotesByID {
		if q.RFQID == rfqID && q.ID != keepQuoteID &&
			(q.Status == entities.QuoteStatusPending || q.Status == entities.QuoteStatusSubmitted) {
			q.Status = entities.QuoteStatusRejected
		}
	}
	return nil
}

func (f *fakeRFQRepo) ListProductsByIDs(_ context.Context, ids []string) ([]*entities.MatchedProduct, error) {
	out := []*entities.MatchedProduct{}
	for _, p := range f.products {
		for _, id := range ids {
			if p.ID == id {
				out = append(out, p)
			}
		}
	}
	return out, nil
}

func (f *fakeRFQRepo) ListSuppliersByIDs(_ context.Context, ids []string) ([]*entities.SupplierContact, error) {
	out := []*entities.SupplierContact{}
	for _, s := range f.suppliers {
		for _, id := range ids {
			if s.ID == id {
				out = append(out, s)
			}
		}
	}
	return out, nil
}

type fakeIssuer struct {
	calls [][2]string
	fail  bool
}

func (f *fakeIssuer) IssueMagicLinkToken(_ context.Context, rfqID, supplierID string) (string, error) {
	f.calls = append(f.calls, [2]string{rfqID, supplierID})
	if f.fail {
		return "", context.DeadlineExceeded
	}
	return "token-" + supplierID, nil
}

type fakePublisher struct {
	created   []messaging.RFQCreatedEvent
	submitted []messaging.QuoteSubmittedEvent
	accepted  []messaging.QuoteAcceptedEvent
}

func (f *fakePublisher) PublishRFQCreated(event messaging.RFQCreatedEvent) error {
	f.created = append(f.created, event)
	return nil
}

func (f *fakePublisher) PublishQuoteSubmitted(event messaging.QuoteSubmittedEvent) error {
	f.submitted = append(f.submitted, event)
	return nil
}

func (f *fakePublisher) PublishQuoteAccepted(event messaging.QuoteAcceptedEvent) error {
	f.accepted = append(f.accepted, event)
	return nil
}

func newTestRFQService(repo *fakeRFQRepo, issuer MagicLinkIssuer, publisher RFQEventPublisher) *RFQService {
	svc := NewRFQService(repo, issuer, publisher, "https://fiberlane.dev/", logger.New("error"))
	svc.now = func() time.Time { return time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC) }
	return svc
}

func seededRepo() *fakeRFQRepo {
	repo := newFakeRFQRepo()
	repo.products = []*entities.MatchedProduct{
		{ID: "11111111-1111-1111-1111-111111111111", SupplierID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", SKU: "QSFP28-LR4", Name: "100G QSFP28 LR4"},
		{ID: "22222222-2222-2222-2222-222222222222", SupplierID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", SKU: "QSFP28-ER4", Name: "100G QSFP28 ER4"},
		{ID: "33333333-3333-3333-3333-333333333333", SupplierID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", SKU: "SFP-10G-LR", Name: "10G SFP+ LR"},
	}
	repo.suppliers = []*entities.SupplierContact{
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: "InnoLight", City: "Suzhou", ContactEmail: "sales@innolight.example"},
		{ID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "Eoptolink", City: "Chengdu", ContactEmail: "rfq@eoptolink.example"},
	}
	return repo
}

func createReq() *dto.CreateRFQRequest {
	return &dto.CreateRFQRequest{
		BuyerID:      "buyer-1",
		BuyerEmail:   "eng@acme-corp.com",
		BuyerCompany: "ACME CORP",
		QueryText:    "100G QSFP28 LR4, Cisco compatible",
		MatchedProductIDs: []string{
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
			"33333333-3333-3333-3333-333333333333",
		},
		Qty:        100,
		TargetDate: "2026-06-30",
	}
}

func TestCreateRFQGroupsQuotesBySupplier(t *testing.T) {
	repo := seededRepo()
	issuer := &fakeIssuer{}
	publisher := &fakePublisher{}
	svc := newTestRFQService(repo, issuer, publisher)

	rfq, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ: %v", err)
	}

	if want := "RFQ-20260609-0001-SZX"; rfq.ID != want {
		t.Errorf("rfq id = %q, want %q", rfq.ID, want)
	}
	if rfq.Status != entities.RFQStatusOpen {
		t.Errorf("status = %q, want open", rfq.Status)
	}

	// Three products across two suppliers -> two pending quotes.
	quotes, _ := repo.ListQuotesForRFQ(context.Background(), rfq.ID)
	if len(quotes) != 2 {
		t.Fatalf("pending quotes = %d, want 2", len(quotes))
	}
	for _, quote := range quotes {
		if quote.Status != entities.QuoteStatusPending {
			t.Errorf("quote %s status = %q, want pending", quote.ID, quote.Status)
		}
		if !strings.HasPrefix(quote.ID, "QUOTE-20260609-") {
			t.Errorf("quote id %q does not follow QUOTE-YYYYMMDD-NNNN-SZX", quote.ID)
		}
	}

	if len(issuer.calls) != 2 {
		t.Fatalf("magic link mints = %d, want 2", len(issuer.calls))
	}

	if len(publisher.created) != 1 {
		t.Fatalf("rfq.created events = %d, want 1", len(publisher.created))
	}
	event := publisher.created[0]
	if event.BuyerEmail != "eng@acme-corp.com" || event.BuyerCompany != "ACME CORP" {
		t.Errorf("event buyer fields = %q/%q", event.BuyerEmail, event.BuyerCompany)
	}
	if len(event.Suppliers) != 2 {
		t.Fatalf("event suppliers = %d, want 2", len(event.Suppliers))
	}
	for _, invite := range event.Suppliers {
		if invite.ContactEmail == "" {
			t.Errorf("invite %s missing contact email", invite.SupplierID)
		}
		wantURL := "https://fiberlane.dev/q/token-" + invite.SupplierID
		if invite.MagicLinkURL != wantURL {
			t.Errorf("invite url = %q, want %q", invite.MagicLinkURL, wantURL)
		}
	}
}

func TestCreateRFQSurvivesMagicLinkFailure(t *testing.T) {
	repo := seededRepo()
	publisher := &fakePublisher{}
	svc := newTestRFQService(repo, &fakeIssuer{fail: true}, publisher)

	rfq, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ should not fail when minting fails: %v", err)
	}
	if len(publisher.created) != 1 {
		t.Fatalf("rfq.created events = %d, want 1", len(publisher.created))
	}
	for _, invite := range publisher.created[0].Suppliers {
		if invite.MagicLinkURL != "" {
			t.Errorf("invite url should be empty on mint failure, got %q", invite.MagicLinkURL)
		}
	}
	if _, err := repo.GetRFQByID(context.Background(), rfq.ID); err != nil {
		t.Errorf("rfq should still be persisted: %v", err)
	}
}

func TestCreateRFQRequiresProducts(t *testing.T) {
	svc := newTestRFQService(seededRepo(), &fakeIssuer{}, &fakePublisher{})
	req := createReq()
	req.MatchedProductIDs = nil
	if _, err := svc.CreateRFQ(context.Background(), req); err != appErrors.ErrNoMatchedProducts {
		t.Errorf("err = %v, want ErrNoMatchedProducts", err)
	}
}

func TestGetRFQEnforcesOwnership(t *testing.T) {
	repo := seededRepo()
	svc := newTestRFQService(repo, &fakeIssuer{}, &fakePublisher{})
	rfq, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ: %v", err)
	}

	if _, err := svc.GetRFQ(context.Background(), rfq.ID, "buyer-1"); err != nil {
		t.Errorf("owner read failed: %v", err)
	}
	if _, err := svc.GetRFQ(context.Background(), rfq.ID, "intruder"); err != appErrors.ErrUnauthorizedAccess {
		t.Errorf("err = %v, want ErrUnauthorizedAccess", err)
	}
}

func TestSubmitQuoteFlow(t *testing.T) {
	repo := seededRepo()
	publisher := &fakePublisher{}
	svc := newTestRFQService(repo, &fakeIssuer{}, publisher)
	rfq, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ: %v", err)
	}

	quote, err := svc.SubmitQuote(context.Background(), &dto.SubmitQuoteRequest{
		RFQID:        rfq.ID,
		SupplierID:   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		PriceUSD:     185.50,
		LeadTimeDays: 12,
	})
	if err != nil {
		t.Fatalf("SubmitQuote: %v", err)
	}
	if quote.Status != entities.QuoteStatusSubmitted {
		t.Errorf("quote status = %q, want submitted", quote.Status)
	}

	updated, _ := repo.GetRFQByID(context.Background(), rfq.ID)
	if updated.Status != entities.RFQStatusQuoted {
		t.Errorf("rfq status = %q, want quoted", updated.Status)
	}

	if len(publisher.submitted) != 1 {
		t.Fatalf("quote.submitted events = %d, want 1", len(publisher.submitted))
	}
	event := publisher.submitted[0]
	if event.BuyerID != "buyer-1" || event.BuyerEmail != "eng@acme-corp.com" || event.SupplierName != "InnoLight" {
		t.Errorf("event = %+v missing buyer/supplier context", event)
	}

	// Re-submission hits the already-submitted guard.
	if _, err := svc.SubmitQuote(context.Background(), &dto.SubmitQuoteRequest{
		RFQID:        rfq.ID,
		SupplierID:   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		PriceUSD:     180,
		LeadTimeDays: 10,
	}); err != appErrors.ErrQuoteAlreadyExists {
		t.Errorf("err = %v, want ErrQuoteAlreadyExists", err)
	}

	// Uninvited supplier gets NotFound.
	if _, err := svc.SubmitQuote(context.Background(), &dto.SubmitQuoteRequest{
		RFQID:        rfq.ID,
		SupplierID:   "cccccccc-cccc-cccc-cccc-cccccccccccc",
		PriceUSD:     180,
		LeadTimeDays: 10,
	}); err != appErrors.ErrQuoteNotFound {
		t.Errorf("err = %v, want ErrQuoteNotFound", err)
	}
}

func TestSubmitManufacturerQuoteHappyPath(t *testing.T) {
	repo := seededRepo()
	publisher := &fakePublisher{}
	svc := newTestRFQService(repo, &fakeIssuer{}, publisher)
	rfq, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ: %v", err)
	}

	quote, err := svc.SubmitManufacturerQuote(context.Background(), &dto.SubmitManufacturerQuoteRequest{
		RFQID:          rfq.ID,
		ManufacturerID: "dddddddd-dddd-dddd-dddd-dddddddddddd",
		PriceUSD:       210.00,
		LeadTimeDays:   10,
	})
	if err != nil {
		t.Fatalf("SubmitManufacturerQuote: %v", err)
	}
	if quote.Status != entities.QuoteStatusSubmitted {
		t.Errorf("quote status = %q, want submitted", quote.Status)
	}
	if quote.ManufacturerID != "dddddddd-dddd-dddd-dddd-dddddddddddd" {
		t.Errorf("manufacturer_id = %q", quote.ManufacturerID)
	}
	if quote.SupplierID != "" {
		t.Errorf("supplier_id should be empty for manufacturer quote, got %q", quote.SupplierID)
	}

	updated, _ := repo.GetRFQByID(context.Background(), rfq.ID)
	if updated.Status != entities.RFQStatusQuoted {
		t.Errorf("rfq status = %q, want quoted", updated.Status)
	}

	if len(publisher.submitted) != 1 {
		t.Fatalf("quote.submitted events = %d, want 1", len(publisher.submitted))
	}
	if publisher.submitted[0].SupplierID != "" {
		t.Errorf("supplier_id in event should be empty for manufacturer quote")
	}
}

func TestSubmitManufacturerQuoteValidation(t *testing.T) {
	svc := newTestRFQService(seededRepo(), &fakeIssuer{}, &fakePublisher{})
	for _, tc := range []struct {
		name string
		req  dto.SubmitManufacturerQuoteRequest
	}{
		{"empty rfq_id", dto.SubmitManufacturerQuoteRequest{ManufacturerID: "mid", PriceUSD: 10, LeadTimeDays: 1}},
		{"empty manufacturer_id", dto.SubmitManufacturerQuoteRequest{RFQID: "r1", PriceUSD: 10, LeadTimeDays: 1}},
		{"zero price", dto.SubmitManufacturerQuoteRequest{RFQID: "r1", ManufacturerID: "mid", LeadTimeDays: 1}},
		{"zero lead_time", dto.SubmitManufacturerQuoteRequest{RFQID: "r1", ManufacturerID: "mid", PriceUSD: 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req
			if _, err := svc.SubmitManufacturerQuote(context.Background(), &req); err != appErrors.ErrInvalidRequest {
				t.Errorf("err = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestAcceptQuoteHappyPath(t *testing.T) {
	repo := seededRepo()
	publisher := &fakePublisher{}
	svc := newTestRFQService(repo, &fakeIssuer{}, publisher)

	rfq, _ := svc.CreateRFQ(context.Background(), createReq())
	// supplier submits a quote
	q, err := svc.SubmitQuote(context.Background(), &dto.SubmitQuoteRequest{
		RFQID:        rfq.ID,
		SupplierID:   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		PriceUSD:     185.50,
		LeadTimeDays: 12,
	})
	if err != nil {
		t.Fatalf("SubmitQuote: %v", err)
	}

	accepted, err := svc.AcceptQuote(context.Background(), rfq.ID, q.ID, "buyer-1")
	if err != nil {
		t.Fatalf("AcceptQuote: %v", err)
	}
	if accepted.Status != entities.QuoteStatusAccepted {
		t.Errorf("quote status = %q, want accepted", accepted.Status)
	}

	updatedRFQ, _ := repo.GetRFQByID(context.Background(), rfq.ID)
	if updatedRFQ.Status != entities.RFQStatusAccepted {
		t.Errorf("rfq status = %q, want accepted", updatedRFQ.Status)
	}

	if len(publisher.accepted) != 1 {
		t.Fatalf("quote.accepted events = %d, want 1", len(publisher.accepted))
	}
	ev := publisher.accepted[0]
	if ev.QuoteID != q.ID || ev.RFQID != rfq.ID || ev.BuyerID != "buyer-1" {
		t.Errorf("event = %+v, unexpected fields", ev)
	}
	if ev.PriceUSD != 185.50 {
		t.Errorf("event price = %v, want 185.50", ev.PriceUSD)
	}

	// Other pending quotes are rejected.
	quotes, _ := repo.ListQuotesForRFQ(context.Background(), rfq.ID)
	for _, qt := range quotes {
		if qt.ID == q.ID {
			continue
		}
		if qt.Status != entities.QuoteStatusRejected {
			t.Errorf("other quote %s status = %q, want rejected", qt.ID, qt.Status)
		}
	}
}

func TestAcceptQuoteWrongActor(t *testing.T) {
	repo := seededRepo()
	svc := newTestRFQService(repo, &fakeIssuer{}, &fakePublisher{})
	rfq, _ := svc.CreateRFQ(context.Background(), createReq())
	q, _ := svc.SubmitQuote(context.Background(), &dto.SubmitQuoteRequest{
		RFQID:        rfq.ID,
		SupplierID:   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		PriceUSD:     185.50,
		LeadTimeDays: 12,
	})

	if _, err := svc.AcceptQuote(context.Background(), rfq.ID, q.ID, "intruder"); err != appErrors.ErrUnauthorizedAccess {
		t.Errorf("err = %v, want ErrUnauthorizedAccess", err)
	}
}

func TestAcceptQuoteNonSubmittedQuote(t *testing.T) {
	repo := seededRepo()
	svc := newTestRFQService(repo, &fakeIssuer{}, &fakePublisher{})
	rfq, _ := svc.CreateRFQ(context.Background(), createReq())

	// pending quote (not yet submitted) cannot be accepted
	quotes, _ := repo.ListQuotesForRFQ(context.Background(), rfq.ID)
	pendingQuote := quotes[0]

	if _, err := svc.AcceptQuote(context.Background(), rfq.ID, pendingQuote.ID, "buyer-1"); err != appErrors.ErrInvalidRequest {
		t.Errorf("err = %v, want ErrInvalidRequest", err)
	}
}

func TestGetRFQForSupplierBlanksBuyerEmail(t *testing.T) {
	repo := seededRepo()
	svc := newTestRFQService(repo, &fakeIssuer{}, &fakePublisher{})
	created, err := svc.CreateRFQ(context.Background(), createReq())
	if err != nil {
		t.Fatalf("CreateRFQ: %v", err)
	}

	rfq, quote, supplierName, err := svc.GetRFQForSupplier(context.Background(), created.ID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("GetRFQForSupplier: %v", err)
	}
	if rfq.BuyerEmail != "" {
		t.Errorf("buyer email leaked to supplier view: %q", rfq.BuyerEmail)
	}
	if quote == nil || quote.Status != entities.QuoteStatusPending {
		t.Errorf("expected pending quote in supplier view, got %+v", quote)
	}
	if supplierName != "InnoLight" {
		t.Errorf("supplier name = %q", supplierName)
	}

	// Supplier that was never invited must not see the RFQ.
	if _, _, _, err := svc.GetRFQForSupplier(context.Background(), created.ID, "cccccccc-cccc-cccc-cccc-cccccccccccc"); err != appErrors.ErrQuoteNotFound {
		t.Errorf("err = %v, want ErrQuoteNotFound", err)
	}
}
