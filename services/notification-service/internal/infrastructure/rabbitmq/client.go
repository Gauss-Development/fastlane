package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"notification-service/internal/config"
	"notification-service/pkg/logger"
	"time"
)

type Client struct {
	config     config.RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	logger     *logger.Logger
	done       chan error
}

type MessageHandler func(string, []byte) error

func NewClient(cfg config.RabbitMQConfig, logger *logger.Logger) *Client {
	return &Client{
		config: cfg,
		logger: logger,
		done:   make(chan error),
	}
}

func (c *Client) Connect() error {
	var err error

	c.connection, err = amqp.Dial(c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbit: %w", err)
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = c.channel.Qos(c.config.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set Qos: %w", err)
	}

	err = c.channel.ExchangeDeclare(
		c.config.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = c.channel.ExchangeDeclare(
		c.config.DLXName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead-letter exchange: %w", err)
	}

	queue, err := c.channel.QueueDeclare(
		c.config.QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    c.config.DLXName,
			"x-dead-letter-routing-key": c.config.DLQRoutingKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, key := range c.config.RoutingKeys() {
		if err := c.channel.QueueBind(
			queue.Name,
			key,
			c.config.ExchangeName,
			false,
			nil); err != nil {
			return fmt.Errorf("failed to bind queue to %q: %w", key, err)
		}
	}

	dlq, err := c.channel.QueueDeclare(
		c.config.DLQName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead-letter queue: %w", err)
	}

	err = c.channel.QueueBind(
		dlq.Name,
		c.config.DLQRoutingKey,
		c.config.DLXName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead-letter queue: %w", err)
	}

	c.logger.Info("Connected to rabbit successfully")
	return nil
}

func (c *Client) StartConsuming(handler MessageHandler) error {
	msgs, err := c.channel.Consume(
		c.config.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to reg consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			c.processMessages(d, handler)
		}
		// msgs is closed when the AMQP connection/channel drops; without this
		// the consumer goroutine would exit silently and emails would stop.
		c.logger.Warn("rabbit consumer stopped: delivery channel closed")
	}()

	c.logger.Info("Start consuming messages from rabbit")
	return nil
}

// NotifyClose returns a channel that fires when the underlying AMQP channel
// closes (e.g. broker restart / connection drop). Callers should treat this as
// fatal and let the orchestrator restart the process to reconnect from scratch.
func (c *Client) NotifyClose() chan *amqp.Error {
	return c.channel.NotifyClose(make(chan *amqp.Error))
}

func (c *Client) processMessages(delivery amqp.Delivery, handler MessageHandler) {
	var err error
	retries := 0

	for retries <= c.config.MaxRetries {
		err = handler(delivery.RoutingKey, delivery.Body)
		if err == nil {
			if ackErr := delivery.Ack(false); ackErr != nil {
				c.logger.Error(fmt.Sprintf("failed to ack message: %v", ackErr))
			}
			return
		}

		retries++
		c.logger.Warn(fmt.Sprintf("message processing failed (attempt %d/%d): %v",
			retries, c.config.MaxRetries+1, err))

		if retries <= c.config.MaxRetries {
			time.Sleep(time.Duration(retries) * time.Second)
		}
	}

	// Reject message if more than retries
	c.logger.Error(fmt.Sprintf("message processing failed after %d atttmps, reject message", c.config.MaxRetries+1))
	if nackErr := delivery.Nack(false, false); nackErr != nil {
		c.logger.Error(fmt.Sprintf("failed to dead-letter message: %v", nackErr))
	}

}

func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error(fmt.Sprintf("failed to close channel %v", err))
		}

		if c.connection != nil {
			if err := c.connection.Close(); err != nil {
				c.logger.Error(fmt.Sprintf("failed to close connection: %v", err))
			}
		}
	}
	c.logger.Info("rabbit connection closed")
	return nil
}

func (c *Client) IsConnected() bool {
	return c.connection != nil && !c.connection.IsClosed()
}
