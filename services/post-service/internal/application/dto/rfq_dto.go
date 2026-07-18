package dto

import "encoding/json"

type CreateRFQRequest struct {
	BuyerID           string          `json:"buyer_id"`
	BuyerEmail        string          `json:"buyer_email"`
	BuyerCompany      string          `json:"buyer_company"`
	QueryText         string          `json:"query_text"`
	ParsedSpecs       json.RawMessage `json:"parsed_specs"`
	MatchedProductIDs []string        `json:"matched_product_ids"`
	Qty               int32           `json:"qty"`
	TargetDate        string          `json:"target_date"`
	ShippingAddress   string          `json:"shipping_address"`
	Notes             string          `json:"notes"`
	ProjectID         string          `json:"project_id"`
}

type ListRFQsRequest struct {
	BuyerID string `json:"buyer_id"`
	Status  string `json:"status"`
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
}

type SubmitQuoteRequest struct {
	RFQID         string  `json:"rfq_id"`
	SupplierID    string  `json:"supplier_id"`
	PriceUSD      float64 `json:"price_usd"`
	LeadTimeDays  int32   `json:"lead_time_days"`
	ValidityDate  string  `json:"validity_date"`
	SupplierNotes string  `json:"supplier_notes"`
}

type SubmitManufacturerQuoteRequest struct {
	RFQID          string  `json:"rfq_id"`
	ManufacturerID string  `json:"manufacturer_id"`
	ProductID      string  `json:"product_id"`
	PriceUSD       float64 `json:"price_usd"`
	LeadTimeDays   int32   `json:"lead_time_days"`
	ValidityDate   string  `json:"validity_date"`
	SupplierNotes  string  `json:"supplier_notes"`
}
