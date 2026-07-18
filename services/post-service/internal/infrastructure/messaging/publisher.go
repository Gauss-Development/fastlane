package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"post-service/pkg/logger"
	"sync"
	"time"
)

type EventPublisher struct {
	mu           sync.RWMutex
	connection   *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	logger       *logger.Logger
}

func NewEventPublisher(rabbitMQURL, exchangeName string, logger *logger.Logger) (*EventPublisher, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	publisher := &EventPublisher{
		connection:   conn,
		channel:      ch,
		exchangeName: exchangeName,
		logger:       logger,
	}

	// Monitor connection
	go publisher.monitorConnection()

	logger.Info("Event publisher initialized successfully")
	return publisher, nil
}

func (p *EventPublisher) publishEvent(routingKey string, event interface{}) error {
	p.mu.RLock()
	ch := p.channel
	p.mu.RUnlock()
	if ch == nil {
		return fmt.Errorf("publisher channel is not available")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = ch.Publish(
		p.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			MessageId:    fmt.Sprintf("%s-%d", routingKey, time.Now().UnixNano()),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info(fmt.Sprintf("Published event: %s with %d bytes", routingKey, len(body)))
	return nil
}

func (p *EventPublisher) monitorConnection() {
	p.mu.RLock()
	conn := p.connection
	ch := p.channel
	p.mu.RUnlock()
	if conn == nil || ch == nil {
		return
	}
	for {
		select {
		case err := <-conn.NotifyClose(make(chan *amqp.Error)):
			if err != nil {
				p.logger.Error(fmt.Sprintf("RabbitMQ connection closed: %v", err))
				return
			}
		case err := <-ch.NotifyClose(make(chan *amqp.Error)):
			if err != nil {
				p.logger.Error(fmt.Sprintf("RabbitMQ channel closed: %v", err))
				return
			}
		}
	}
}

func (p *EventPublisher) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connection != nil && !p.connection.IsClosed() && p.channel != nil
}

func (p *EventPublisher) Reconnect(rabbitMQURL string) error {
	p.logger.Info("Attempting to reconnect to RabbitMQ...")

	// Create new connection
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel during reconnect: %w", err)
	}

	// Redeclare exchange
	err = ch.ExchangeDeclare(
		p.exchangeName, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to redeclare exchange during reconnect: %w", err)
	}

	// Swap in the new connection/channel under the write lock, closing the old.
	p.mu.Lock()
	p.closeLocked()
	p.connection = conn
	p.channel = ch
	p.mu.Unlock()

	// Restart monitoring
	go p.monitorConnection()

	p.logger.Info("Successfully reconnected to RabbitMQ")
	return nil
}

func (p *EventPublisher) Close() error {
	p.logger.Info("Closing event publisher...")

	p.mu.Lock()
	p.closeLocked()
	p.mu.Unlock()

	p.logger.Info("Event publisher closed")
	return nil
}

// closeLocked closes and nils the channel/connection. Caller must hold p.mu.
func (p *EventPublisher) closeLocked() {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to close channel: %v", err))
		}
		p.channel = nil
	}

	if p.connection != nil {
		if err := p.connection.Close(); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to close connection: %v", err))
		}
		p.connection = nil
	}
}

func (p *EventPublisher) HealthCheck() error {
	if !p.IsConnected() {
		return fmt.Errorf("event publisher is not connected")
	}
	return nil
}
