package rmqrpc

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeKind = "topic"
)

// Config -.
type Config struct {
	URL      string
	WaitTime time.Duration
	Attempts int
}

// Connection -.
type Connection struct {
	ConsumerExchange string
	QueueName        string
	AppName          string
	RoutingKey       string
	RKQueueName      string
	Config
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Delivery   <-chan amqp.Delivery
}

// New -.
func New(appName, consumerExchange, queueName, rk string, cfg Config) *Connection {
	conn := &Connection{
		ConsumerExchange: consumerExchange,
		QueueName:        queueName,
		AppName:          appName,
		Config:           cfg,
		RKQueueName:      rk,
	}

	return conn
}

// AttemptConnect -.
func (c *Connection) AttemptConnect() error {
	var err error
	for i := c.Attempts; i > 0; i-- {
		if err = c.connect(); err == nil {
			break
		}

		log.Printf("RabbitMQ is trying to connect, attempts left: %d", i)
		time.Sleep(c.WaitTime)
	}

	if err != nil {
		return fmt.Errorf("rmq_rpc - AttemptConnect - c.connect: %w", err)
	}

	return nil
}

// SetRoutingKey - set routing key.
func (c *Connection) SetRoutingKey(rk string) {
	if rk != "" {
		c.RoutingKey = rk
	}
}

// getRoutingKey - gets routing key depends of set.
func (c *Connection) getRoutingKey() string {
	if c.RoutingKey == "" {
		return "#." + c.AppName + "." + c.RKQueueName
	}
	return c.RoutingKey
}

func (c *Connection) connect() error {
	var err error
	c.Connection, err = amqp.Dial(c.URL)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	c.Channel, err = c.Connection.Channel()
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	err = c.Channel.ExchangeDeclare(
		c.ConsumerExchange,
		exchangeKind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	queue, err := c.Channel.QueueDeclare(
		c.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueDeclare: %w", err)
	}

	err = c.Channel.QueueBind(
		queue.Name,
		c.getRoutingKey(),
		c.ConsumerExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueBind: %w", err)
	}

	c.Delivery, err = c.Channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.Consume: %w", err)
	}

	return nil
}
