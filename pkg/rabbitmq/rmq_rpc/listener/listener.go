package listener

import (
	"fmt"
	"time"

	"test_go/pkg/rabbitmq/utils"

	amqp "github.com/rabbitmq/amqp091-go"
	"test_go/pkg/logger"
	rmqrpc "test_go/pkg/rabbitmq/rmq_rpc"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 5 * time.Second

	_routingKeyForAllMessages = "#"
)

// CallHandler -.
type CallHandler func(*amqp.Delivery) (interface{}, error)

// Listener -.
type Listener struct {
	conn    *rmqrpc.Connection
	error   chan error
	stop    chan struct{}
	handler []CallHandler
	timeout time.Duration

	logger logger.Interface
}

// New - constructor with raw handler for rmq listener.
func New(
	url, exchange, appName string,
	hdl []CallHandler,
	l logger.Interface) (*Listener, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	queue := utils.GetListenerQueueName(exchange, appName)
	if queue == "" {
		return nil, fmt.Errorf("rmq_rpc listener - New - wrong queue name provided")
	}

	s := &Listener{
		conn:    rmqrpc.New(appName, exchange, queue, queue, cfg),
		error:   make(chan error),
		stop:    make(chan struct{}),
		handler: hdl,
		timeout: _defaultTimeout,
		logger:  l,
	}

	// Set # routing key for raw handler
	s.conn.SetRoutingKey(_routingKeyForAllMessages)

	err := s.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc listener - New - s.conn.AttemptConnect: %w", err)
	}

	go s.consumer()

	return s, nil
}

func (s *Listener) consumer() {
	for {
		select {
		case <-s.stop:
			return
		case d, opened := <-s.conn.Delivery:
			if !opened {
				s.reconnect()

				return
			}

			_ = d.Ack(false)

			s.serveCall(&d)
		}
	}
}

func (s *Listener) serveCall(d *amqp.Delivery) {
	// Process the message using the handler
	_, err := s.handler[0](d)

	if err != nil {
		s.logger.Error(err, "rmq_rpc listener - Listener - serveCall - handler")
	}
}

func (s *Listener) reconnect() {
	close(s.stop)

	err := s.conn.AttemptConnect()
	if err != nil {
		s.error <- err
		close(s.error)

		return
	}

	s.stop = make(chan struct{})

	go s.consumer()
}

// Notify -.
func (s *Listener) Notify() <-chan error {
	return s.error
}

// Shutdown -.
func (s *Listener) Shutdown() error {
	select {
	case <-s.error:
		return nil
	default:
	}

	close(s.stop)
	time.Sleep(s.timeout)

	err := s.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc listener - Listener - Shutdown - s.Connection.Close: %w", err)
	}

	return nil
}
