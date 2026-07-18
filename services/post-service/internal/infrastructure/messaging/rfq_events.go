package messaging

import "time"

// RFQSupplierInvite carries everything notification-service needs to email
// one supplier: contact, part context, and the pre-minted magic link.
type RFQSupplierInvite struct {
	SupplierID   string `json:"supplier_id"`
	Name         string `json:"name"`
	NameZh       string `json:"name_zh"`
	City         string `json:"city"`
	ContactEmail string `json:"contact_email"`
	MagicLinkURL string `json:"magic_link_url"`
	ProductSKU   string `json:"product_sku"`
	ProductName  string `json:"product_name"`
}

type RFQCreatedEvent struct {
	RFQID           string              `json:"rfq_id"`
	BuyerID         string              `json:"buyer_id"`
	BuyerEmail      string              `json:"buyer_email"`
	BuyerCompany    string              `json:"buyer_company"`
	QueryText       string              `json:"query_text"`
	PartSummary     string              `json:"part_summary"`
	Qty             int32               `json:"qty"`
	TargetDate      string              `json:"target_date"`
	ShippingAddress string              `json:"shipping_address"`
	Notes           string              `json:"notes"`
	Suppliers       []RFQSupplierInvite `json:"suppliers"`
	CreatedAt       time.Time           `json:"created_at"`
}

type QuoteSubmittedEvent struct {
	QuoteID      string    `json:"quote_id"`
	RFQID        string    `json:"rfq_id"`
	BuyerID      string    `json:"buyer_id"`
	BuyerEmail   string    `json:"buyer_email"`
	BuyerCompany string    `json:"buyer_company"`
	QueryText    string    `json:"query_text"`
	SupplierID   string    `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	PriceUSD     float64   `json:"price_usd"`
	LeadTimeDays int32     `json:"lead_time_days"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

type QuoteAcceptedEvent struct {
	RFQID           string    `json:"rfq_id"`
	QuoteID         string    `json:"quote_id"`
	BuyerID         string    `json:"buyer_id"`
	BuyerEmail      string    `json:"buyer_email"`
	BuyerCompany    string    `json:"buyer_company"`
	QueryText       string    `json:"query_text"`
	SupplierID      string    `json:"supplier_id"`      // set for seed-supplier quotes; "" for manufacturer quotes
	ManufacturerID  string    `json:"manufacturer_id"`  // set for manufacturer quotes; "" otherwise
	ProductID       string    `json:"product_id"`
	PriceUSD        float64   `json:"price_usd"`
	Qty             int32     `json:"qty"`
	ShippingAddress string    `json:"shipping_address"`
	AcceptedAt      time.Time `json:"accepted_at"`
}

func (p *EventPublisher) PublishRFQCreated(event RFQCreatedEvent) error {
	return p.publishEvent("rfq.created", event)
}

func (p *EventPublisher) PublishQuoteSubmitted(event QuoteSubmittedEvent) error {
	return p.publishEvent("quote.submitted", event)
}

func (p *EventPublisher) PublishQuoteAccepted(event QuoteAcceptedEvent) error {
	return p.publishEvent("quote.accepted", event)
}
