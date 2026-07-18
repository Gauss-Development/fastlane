package entities

import "time"

type Order struct {
	ID                 string
	BuyerID            string
	SupplierID         string
	QuoteID            string
	RFQID              string
	Status             string
	PaymentStatus      string
	QCStatus           string
	TotalUSD           float64
	ShippingAddress    string
	ShippingCity       string
	ShippingCountry    string
	WarrantyUntil      string // ISO-8601 date or empty
	CancelledAt        *time.Time
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrderEvent struct {
	ID          string
	OrderID     string
	EventType   string
	FromStatus  string
	ToStatus    string
	ActorID     string
	ActorType   string
	OccurredAt  time.Time
	OccurredTZ  string
	Location    string
	Payload     []byte // raw jsonb
	Documents   []byte // raw jsonb
	Notes       string
	CreatedAt   time.Time
}

// allowedTransitions defines the valid state machine edges.
// Happy path: pending_payment→paid→in_production→ready_for_qc→qc_in_progress→(qc_passed|qc_failed)
// qc_passed→shipped_from_cn→in_transit→out_for_delivery→delivered→completed
// Cancelled from pre-shipment states. Disputed from post-payment states.
var allowedTransitions = map[string][]string{
	"draft":            {"pending_payment", "cancelled"},
	"pending_payment":  {"paid", "cancelled"},
	"paid":             {"in_production", "cancelled", "refunded"},
	"in_production":    {"ready_for_qc", "cancelled"},
	"ready_for_qc":     {"qc_in_progress"},
	"qc_in_progress":   {"qc_passed", "qc_failed"},
	"qc_failed":        {"qc_in_progress", "cancelled"},
	"qc_passed":        {"shipped_from_cn"},
	"shipped_from_cn":  {"in_transit"},
	"in_transit":       {"out_for_delivery", "disputed"},
	"out_for_delivery": {"delivered"},
	"delivered":        {"completed", "disputed"},
	"completed":        {},
	"cancelled":        {},
	"refunded":         {},
	"disputed":         {"completed", "refunded"},
}

// CanTransition returns true if moving from→to is a valid state machine edge.
func CanTransition(from, to string) bool {
	nexts, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, n := range nexts {
		if n == to {
			return true
		}
	}
	return false
}
