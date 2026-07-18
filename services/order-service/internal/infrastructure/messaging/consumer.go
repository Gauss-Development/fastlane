package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"order-service/internal/application/services"
	"order-service/internal/config"
	"order-service/pkg/logger"
)

const (
	queueName  = "order.quote_accepted"
	routingKey = "quote.accepted"
	dlxSuffix  = ".dlx"
	dlqSuffix  = ".dlq"
)

// QuoteAcceptedEvent is the JSON payload published by post-service on quote.accepted.
type QuoteAcceptedEvent struct {
	RFQID           string    `json:"rfq_id"`
	QuoteID         string    `json:"quote_id"`
	BuyerID         string    `json:"buyer_id"`
	BuyerEmail      string    `json:"buyer_email"`
	BuyerCompany    string    `json:"buyer_company"`
	QueryText       string    `json:"query_text"`
	SupplierID      string    `json:"supplier_id"`
	ManufacturerID  string    `json:"manufacturer_id"`
	ProductID       string    `json:"product_id"`
	PriceUSD        float64   `json:"price_usd"`
	Qty             int32     `json:"qty"`
	ShippingAddress string    `json:"shipping_address"`
	AcceptedAt      time.Time `json:"accepted_at"`
}

// Consumer connects to RabbitMQ and consumes quote.accepted events.
type Consumer struct {
	cfg    config.RabbitMQConfig
	svc    *services.OrderService
	logger *logger.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func NewConsumer(cfg config.RabbitMQConfig, svc *services.OrderService, log *logger.Logger) *Consumer {
	return &Consumer{cfg: cfg, svc: svc, logger: log}
}

// Start connects and begins consuming. Returns nil immediately; consumes in a goroutine.
// If RABBITMQ_URL is empty the consumer is disabled (no-op).
func (c *Consumer) Start(ctx context.Context) error {
	if c.cfg.URL == "" {
		c.logger.Warn("order consumer: RABBITMQ_URL empty, consumer disabled")
		return nil
	}

	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("order consumer: connect: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("order consumer: open channel: %w", err)
	}
	c.ch = ch

	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("order consumer: qos: %w", err)
	}

	// Declare main exchange (shared, must match post-service settings).
	if err := ch.ExchangeDeclare(c.cfg.ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("order consumer: declare exchange: %w", err)
	}

	// Declare DLX.
	dlx := c.cfg.ExchangeName + dlxSuffix
	if err := ch.ExchangeDeclare(dlx, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("order consumer: declare dlx: %w", err)
	}

	// Declare durable queue with DLX.
	q, err := ch.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    dlx,
		"x-dead-letter-routing-key": queueName + ".failed",
	})
	if err != nil {
		return fmt.Errorf("order consumer: declare queue: %w", err)
	}
	if err := ch.QueueBind(q.Name, routingKey, c.cfg.ExchangeName, false, nil); err != nil {
		return fmt.Errorf("order consumer: bind queue: %w", err)
	}

	// DLQ.
	dlq, err := ch.QueueDeclare(queueName+dlqSuffix, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("order consumer: declare dlq: %w", err)
	}
	if err := ch.QueueBind(dlq.Name, queueName+".failed", dlx, false, nil); err != nil {
		return fmt.Errorf("order consumer: bind dlq: %w", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("order consumer: consume: %w", err)
	}

	go func() {
		c.logger.Info("order consumer: started, waiting for quote.accepted")
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("order consumer: context done, stopping")
				return
			case d, ok := <-msgs:
				if !ok {
					c.logger.Warn("order consumer: delivery channel closed")
					return
				}
				c.handle(d)
			}
		}
	}()

	// Watch for connection close and log — container restart policy handles reconnect.
	go func() {
		if closeErr := <-conn.NotifyClose(make(chan *amqp.Error)); closeErr != nil {
			c.logger.Fatal("order consumer: connection closed, exiting for restart: " + closeErr.Error())
		}
	}()

	return nil
}

func (c *Consumer) handle(d amqp.Delivery) {
	var evt QuoteAcceptedEvent
	if err := json.Unmarshal(d.Body, &evt); err != nil {
		// Poison message — nack without requeue so it goes to DLQ.
		c.logger.Error(fmt.Sprintf("order consumer: unmarshal quote.accepted: %v", err))
		_ = d.Nack(false, false)
		return
	}

	if _, err := c.svc.CreateOrderFromQuote(context.Background(), services.QuoteAcceptedEvent(evt)); err != nil {
		c.logger.Error(fmt.Sprintf("order consumer: CreateOrderFromQuote %s: %v", evt.QuoteID, err))
		_ = d.Nack(false, true) // requeue transient errors
		return
	}
	_ = d.Ack(false)
}

func (c *Consumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
