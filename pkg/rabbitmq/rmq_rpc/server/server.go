package server

import (
	"errors"
	"fmt"
	"test_go/pkg/rabbitmq/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"test_go/pkg/logger"
	rmqrpc "test_go/pkg/rabbitmq/rmq_rpc"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 5 * time.Second

	serverName = "server"
	delimeter  = "_"
)

// CallHandler -.
type CallHandler func(*amqp.Delivery) (interface{}, error)

// Server -.
type Server struct {
	conn    *rmqrpc.Connection
	error   chan error
	stop    chan struct{}
	router  map[string]CallHandler
	handler CallHandler

	timeout time.Duration

	logger logger.Interface
}

// New - constructor for rmq server.
func New(
	url, serverExchange, appName string,
	router map[string]CallHandler,
	l logger.Interface,
	queuePrefix string,
	opts ...Option) (*Server, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	if queuePrefix == "" {
		return nil, errors.New("rmq client error - invalid queue prefix")
	}

	//queue := appName + delimeter + queuePrefix + delimeter + serverName
	rk := appName + delimeter + serverName

	s := &Server{
		conn:    rmqrpc.New(appName, serverExchange, rk, rk, cfg),
		error:   make(chan error),
		stop:    make(chan struct{}),
		router:  router,
		timeout: _defaultTimeout,
		logger:  l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	err := s.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc server - NewServer - s.conn.AttemptConnect: %w", err)
	}

	go s.consumer()

	return s, nil
}

func (s *Server) consumer() {
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

func (s *Server) serveCall(d *amqp.Delivery) {
	callHandler, ok := s.router[d.Type]
	if !ok {
		s.publish(d, nil, rmqrpc.ErrBadHandler.Error())
		return
	}

	status := rmqrpc.Success
	data, err := callHandler(d)
	message := rmqrpc.CastToMessageResponse(data)

	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - serveCall - callHandler")

		var messageError rmqrpc.MessageError
		if errors.As(err, &messageError) {
			e := err.(rmqrpc.MessageError)
			message.Error = &e
		} else {
			message.Error = &rmqrpc.MessageError{
				Code:    rmqrpc.Internal,
				Message: err.Error(),
			}
		}
		status = rmqrpc.ErrInternalServer.Error()
	} else {
		message.Error = nil
	}

	body, err := message.Pack()
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - serveCall - json.Marshal")
	}

	s.publish(d, body, status)
}

func (s *Server) publish(d *amqp.Delivery, body []byte, status string) {
	convertedRoutingKey, err := utils.ConvertRoutingKey(d.RoutingKey)
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - convertRoutingKey")
	}

	err = s.conn.Channel.Publish(d.ReplyTo, convertedRoutingKey, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Headers:       d.Headers,
			Type:          status,
			Body:          body,
		})
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - publish - s.conn.Channel.Publish")
	}
}

func (s *Server) reconnect() {
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
func (s *Server) Notify() <-chan error {
	return s.error
}

// Shutdown -.
func (s *Server) Shutdown() error {
	select {
	case <-s.error:
		return nil
	default:
	}

	close(s.stop)
	time.Sleep(s.timeout)

	err := s.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.Connection.Close: %w", err)
	}

	return nil
}
