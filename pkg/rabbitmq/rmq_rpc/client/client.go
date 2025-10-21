package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"test_go/pkg/rabbitmq/utils"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	rmqrpc "test_go/pkg/rabbitmq/rmq_rpc"
)

// ErrConnectionClosed -.
var ErrConnectionClosed = errors.New("rmq_rpc client - Client - RemoteCall - Connection closed")

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 5 * time.Second

	clientName = "client"
	delimeter  = "_"
)

type pendingCall struct {
	done   chan struct{}
	status string
	body   []byte
}

// Client -.
type Client struct {
	conn           *rmqrpc.Connection
	serverExchange string
	appName        string
	error          chan error
	stop           chan struct{}
	headers        map[string]interface{}

	rw    sync.RWMutex
	calls map[string]*pendingCall

	timeout      time.Duration
	onlySendMode bool
}

// New -.
func New(url, serverExchange, clientExchange, appName, queuePrefix string, opts ...Option) (*Client, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	if queuePrefix == "" {
		return nil, errors.New("rmq client error - invalid queue prefix")
	}

	queue := appName + delimeter + queuePrefix + delimeter + clientName
	rk := appName + delimeter + clientName

	c := &Client{
		conn:           rmqrpc.New(appName, clientExchange, queue, rk, cfg),
		serverExchange: serverExchange,
		appName:        appName,
		error:          make(chan error),
		stop:           make(chan struct{}),
		calls:          make(map[string]*pendingCall),
		timeout:        _defaultTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(c)
	}

	err := c.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc client - NewClient - c.conn.AttemptConnect: %w", err)
	}

	go c.consumer()

	return c, nil
}

func (c *Client) publish(ctx context.Context, receiver, corrID, handler string, request interface{}) error {
	var (
		requestBody []byte
		err         error
	)

	if request != nil {
		requestBody, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}

	routingKey, err := utils.ConstructRoutingKey(c.appName, receiver)
	if err != nil {
		return err
	}

	_publishing := amqp.Publishing{
		ContentType:  "application/json",
		Headers:      c.headers,
		Type:         handler,
		Body:         requestBody,
		DeliveryMode: amqp.Persistent,
	}

	if !c.onlySendMode {
		_publishing.CorrelationId = corrID
		_publishing.ReplyTo = c.conn.ConsumerExchange
	}

	err = c.conn.Channel.PublishWithContext(ctx, c.serverExchange, routingKey, false, false, _publishing)
	if err != nil {
		return fmt.Errorf("c.Channel.Publish: %w", err)
	}

	return nil
}

// RemoteCall -.
func (c *Client) RemoteCall(ctx context.Context, receiver, handler string, request, response interface{}) error {
	select {
	case <-c.stop:
		time.Sleep(c.timeout)
		select {
		case <-c.stop:
			return ErrConnectionClosed
		default:
		}
	default:
	}

	corrID := uuid.New().String()

	// checks if request is message request type.
	mr, ok := rmqrpc.CheckAndCastToMessageRequest(request)
	if ok {
		c.headers = mr.Headers
		request = mr.Payload
	}
	// check and update header for listener
	c.checkAndUpdateHeaderForListener(ctx)

	err := c.publish(ctx, receiver, corrID, handler, request)
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - RemoteCall - c.publish: %w", err)
	}

	if !c.onlySendMode {
		call := &pendingCall{done: make(chan struct{})}

		c.addCall(corrID, call)
		defer c.deleteCall(corrID)

		select {
		case <-time.After(c.timeout):
			return rmqrpc.ErrTimeout
		case <-call.done:
		}

		resp := &rmqrpc.MessageResponse{}
		err = resp.Unpack(call.body, &response)
		if err != nil {
			return err
		}

		if call.status == rmqrpc.Success {
			return nil
		}

		if call.status == rmqrpc.ErrBadHandler.Error() {
			return rmqrpc.ErrBadHandler
		}

		if call.status == rmqrpc.ErrInternalServer.Error() {
			return resp.Error
		}
	}

	return nil
}

// RemoteCallWithCustomError - modified RemoteCall method that returns extended error.
func (c *Client) RemoteCallWithCustomError(ctx context.Context, receiver, handler string, request, response interface{}) error {
	select {
	case <-c.stop:
		time.Sleep(c.timeout)
		select {
		case <-c.stop:
			return ErrConnectionClosed
		default:
		}
	default:
	}

	corrID := uuid.New().String()

	// checks if request is message request type.
	mr, ok := rmqrpc.CheckAndCastToMessageRequest(request)
	if ok {
		c.headers = mr.Headers
		request = mr.Payload
	}
	// check and update header for listener
	c.checkAndUpdateHeaderForListener(ctx)

	err := c.publish(ctx, receiver, corrID, handler, request)
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - RemoteCall - c.publish: %w", err)
	}

	if !c.onlySendMode {
		call := &pendingCall{done: make(chan struct{})}

		c.addCall(corrID, call)
		defer c.deleteCall(corrID)

		select {
		case <-time.After(c.timeout):
			return rmqrpc.ErrTimeout
		case <-call.done:
		}

		out := &rmqrpc.MessageResponse{}
		err = json.Unmarshal(call.body, &out)
		if err != nil {
			if out.Error != nil {
				return out.Error
			}
			return fmt.Errorf("rmq_rpc - RemoteCall - json.Unmarshal: %w", err)
		}

		err = out.Unpack(call.body, &response)
		if err != nil {
			if out.Error != nil {
				return out.Error
			}
			return fmt.Errorf("rmq_rpc - RemoteCall - json.Unmarshal: %w", err)
		}

		if call.status == rmqrpc.Success {
			return nil
		}

		if call.status == rmqrpc.ErrBadHandler.Error() {
			return rmqrpc.ErrBadHandler
		}

		if call.status == rmqrpc.ErrInternalServer.Error() {
			return out.Error
		}
	}

	return nil
}

func (c *Client) consumer() {
	for {
		select {
		case <-c.stop:
			return
		case d, opened := <-c.conn.Delivery:
			if !opened {
				c.reconnect()

				return
			}

			_ = d.Ack(false)

			c.getCall(&d)
		}
	}
}

func (c *Client) SwitchClientOnlySendMode(isEnabled bool) {
	c.onlySendMode = isEnabled
}

func (c *Client) reconnect() {
	close(c.stop)

	err := c.conn.AttemptConnect()
	if err != nil {
		c.error <- err
		close(c.error)

		return
	}

	c.stop = make(chan struct{})

	go c.consumer()
}

func (c *Client) getCall(d *amqp.Delivery) {
	c.rw.RLock()
	call, ok := c.calls[d.CorrelationId]
	c.rw.RUnlock()

	if !ok {
		return
	}

	call.status = d.Type
	call.body = d.Body
	close(call.done)
}

func (c *Client) addCall(corrID string, call *pendingCall) {
	c.rw.Lock()
	c.calls[corrID] = call
	c.rw.Unlock()
}

func (c *Client) deleteCall(corrID string) {
	c.rw.Lock()
	delete(c.calls, corrID)
	c.rw.Unlock()
}

// Notify -.
func (c *Client) Notify() <-chan error {
	return c.error
}

// Shutdown -.
func (c *Client) Shutdown() error {
	select {
	case <-c.error:
		return nil
	default:
	}

	close(c.stop)
	time.Sleep(c.timeout)

	err := c.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - Shutdown - c.Connection.Close: %w", err)
	}

	return nil
}

func (c *Client) checkAndUpdateHeaderForListener(ctx context.Context) {
	if !utils.CheckListenerPropertyFromContext(ctx) {
		return
	}

	k := utils.LoggerKey
	v := ctx.Value(k)

	if c.headers != nil {
		c.headers[k] = v
	} else {
		c.headers = map[string]interface{}{
			k: v,
		}
	}
}
