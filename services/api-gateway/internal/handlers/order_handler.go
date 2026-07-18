package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"api-gateway/pkg/logger"
	"api-gateway/pkg/utils"

	orderv1 "github.com/nikitashilov/microblog_grpc/proto/order/v1"
)

// OrderClientAPI is the minimal client surface used by OrderHandler (testable).
type OrderClientAPI interface {
	ListOrders(ctx context.Context, buyerID, orderStatus string, limit, offset int32) (*orderv1.ListOrdersResponse, error)
	GetOrder(ctx context.Context, id, requestingUserID string) (*orderv1.Order, error)
	ListEvents(ctx context.Context, orderID string) (*orderv1.ListEventsResponse, error)
	AppendEvent(ctx context.Context, event *orderv1.OrderEvent) (*orderv1.OrderEvent, error)
}

type OrderHandler struct {
	orderClient OrderClientAPI
	logger      *logger.Logger
}

func NewOrderHandler(orderClient OrderClientAPI, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{orderClient: orderClient, logger: logger}
}

// ListOrders handles GET /api/v1/orders.
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit := parseQueryInt(c, "limit", 20, 1, 100)
	offset := parseQueryInt(c, "offset", 0, 0, 1<<30)

	resp, err := h.orderClient.ListOrders(c.Request.Context(), userID.(string), c.Query("status"), limit, offset)
	if err != nil {
		h.handleOrderError(c, err, "LIST_ORDERS_FAILED", "Failed to list orders")
		return
	}

	orders := make([]map[string]interface{}, 0, len(resp.GetOrders()))
	for _, o := range resp.GetOrders() {
		orders = append(orders, orderToMap(o))
	}
	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", map[string]interface{}{
		"orders": orders,
		"total":  resp.GetTotal(),
		"limit":  limit,
		"offset": offset,
	})
}

// GetOrder handles GET /api/v1/orders/:id.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	o, err := h.orderClient.GetOrder(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		h.handleOrderError(c, err, "GET_ORDER_FAILED", "Failed to retrieve order")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", orderToMap(o))
}

// ListEvents handles GET /api/v1/orders/:id/events.
func (h *OrderHandler) ListEvents(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Authorize: caller must have access to the order.
	if _, err := h.orderClient.GetOrder(c.Request.Context(), c.Param("id"), userID.(string)); err != nil {
		h.handleOrderError(c, err, "GET_ORDER_FAILED", "Failed to retrieve order")
		return
	}

	resp, err := h.orderClient.ListEvents(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleOrderError(c, err, "LIST_EVENTS_FAILED", "Failed to list order events")
		return
	}

	events := make([]map[string]interface{}, 0, len(resp.GetEvents()))
	for _, e := range resp.GetEvents() {
		events = append(events, orderEventToMap(e))
	}
	utils.SuccessResponse(c, http.StatusOK, "Order events retrieved successfully", map[string]interface{}{
		"events": events,
	})
}

type appendEventBody struct {
	ToStatus  string `json:"to_status"`
	EventType string `json:"event_type"`
	Notes     string `json:"notes"`
	Location  string `json:"location"`
}

// AppendEvent handles POST /api/v1/orders/:id/events.
func (h *OrderHandler) AppendEvent(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Authorize: caller must have access to the order.
	if _, err := h.orderClient.GetOrder(c.Request.Context(), c.Param("id"), userID.(string)); err != nil {
		h.handleOrderError(c, err, "GET_ORDER_FAILED", "Failed to retrieve order")
		return
	}

	var body appendEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON")
		return
	}
	if body.EventType == "" || body.ToStatus == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "event_type and to_status are required")
		return
	}

	event, err := h.orderClient.AppendEvent(c.Request.Context(), &orderv1.OrderEvent{
		OrderId:    c.Param("id"),
		EventType:  body.EventType,
		ToStatus:   body.ToStatus,
		ActorId:    userID.(string),
		ActorType:  "buyer",
		OccurredAt: timestamppb.New(time.Now().UTC()),
		Location:   body.Location,
		Notes:      body.Notes,
	})
	if err != nil {
		h.handleOrderError(c, err, "APPEND_EVENT_FAILED", "Failed to append order event")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Order event appended successfully", orderEventToMap(event))
}

func (h *OrderHandler) handleOrderError(c *gin.Context, err error, fallbackCode, fallbackMessage string) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			return
		case codes.NotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", st.Message())
			return
		case codes.PermissionDenied:
			utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", st.Message())
			return
		case codes.Unauthenticated:
			utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", st.Message())
			return
		case codes.Unavailable:
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "ORDER_UNAVAILABLE", "Order service temporarily unavailable")
			return
		}
	}
	h.logger.Error(fallbackMessage + ": " + err.Error())
	utils.ErrorResponse(c, http.StatusInternalServerError, fallbackCode, fallbackMessage)
}

func orderToMap(o *orderv1.Order) map[string]interface{} {
	if o == nil {
		return nil
	}
	return map[string]interface{}{
		"id":               o.GetId(),
		"buyer_id":         o.GetBuyerId(),
		"supplier_id":      o.GetSupplierId(),
		"quote_id":         o.GetQuoteId(),
		"rfq_id":           o.GetRfqId(),
		"status":           o.GetStatus(),
		"payment_status":   o.GetPaymentStatus(),
		"qc_status":        o.GetQcStatus(),
		"total_usd":        o.GetTotalUsd(),
		"shipping_address": o.GetShippingAddress(),
		"shipping_city":    o.GetShippingCity(),
		"shipping_country": o.GetShippingCountry(),
		"warranty_until":   o.GetWarrantyUntil(),
		"created_at":       timestampString(o.GetCreatedAt()),
		"updated_at":       timestampString(o.GetUpdatedAt()),
	}
}

func orderEventToMap(e *orderv1.OrderEvent) map[string]interface{} {
	if e == nil {
		return nil
	}
	return map[string]interface{}{
		"id":          e.GetId(),
		"order_id":    e.GetOrderId(),
		"event_type":  e.GetEventType(),
		"from_status": e.GetFromStatus(),
		"to_status":   e.GetToStatus(),
		"actor_id":    e.GetActorId(),
		"actor_type":  e.GetActorType(),
		"occurred_at": timestampString(e.GetOccurredAt()),
		"location":    e.GetLocation(),
		"notes":       e.GetNotes(),
	}
}
