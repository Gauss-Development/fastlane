package entities

import (
	"encoding/json"
	"time"
)

// RFQ statuses walk open -> quoted -> accepted -> closed.
const (
	RFQStatusOpen     = "open"
	RFQStatusQuoted   = "quoted"
	RFQStatusAccepted = "accepted"
	RFQStatusClosed   = "closed"
)

// Quote statuses: a pending row is created per invited supplier when the RFQ
// is published; submission flips it to submitted.
const (
	QuoteStatusPending   = "pending"
	QuoteStatusSubmitted = "submitted"
	QuoteStatusAccepted  = "accepted"
	QuoteStatusRejected  = "rejected"
)

type RFQ struct {
	ID                string          `json:"id"`
	BuyerID           string          `json:"buyer_id"`
	BuyerEmail        string          `json:"buyer_email"`
	BuyerCompany      string          `json:"buyer_company"`
	QueryText         string          `json:"query_text"`
	ParsedSpecs       json.RawMessage `json:"parsed_specs"`
	MatchedProductIDs []string        `json:"matched_product_ids"`
	Status            string          `json:"status"`
	Qty               int32           `json:"qty"`
	TargetDate        string          `json:"target_date"` // ISO-8601 date, empty when unset
	ShippingAddress   string          `json:"shipping_address"`
	Notes             string          `json:"notes"`
	CreatedAt         time.Time       `json:"created_at"`
}

type Quote struct {
	ID            string    `json:"id"`
	RFQID         string    `json:"rfq_id"`
	SupplierID    string    `json:"supplier_id"`
	ProductID     string    `json:"product_id"`
	PriceUSD      float64   `json:"price_usd"`
	LeadTimeDays  int32     `json:"lead_time_days"`
	ValidityDate  string    `json:"validity_date"` // ISO-8601 date, empty when unset
	SupplierNotes string    `json:"supplier_notes"`
	MatchScore    int32     `json:"match_score"`
	Status        string    `json:"status"`
	SubmittedAt   time.Time `json:"submitted_at"` // zero when not submitted
	CreatedAt     time.Time `json:"created_at"`
}

// MatchedProduct is the slice of the catalog the RFQ flow needs: enough to
// group invitations by supplier and describe the part in supplier emails.
type MatchedProduct struct {
	ID           string  `json:"id"`
	SupplierID   string  `json:"supplier_id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	NameZh       string  `json:"name_zh"`
	Category     string  `json:"category"`
	Specs        []byte  `json:"specs"`
	PriceUSD     float64 `json:"price_usd"`
	MOQ          int32   `json:"moq"`
	LeadTimeDays int32   `json:"lead_time_days"`
}

// SupplierContact is what the notification flow needs about an invited
// supplier; the full supplier profile stays in the catalog queries.
type SupplierContact struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	NameZh       string `json:"name_zh"`
	City         string `json:"city"`
	ContactEmail string `json:"contact_email"`
	Verified     bool   `json:"verified"`
}
