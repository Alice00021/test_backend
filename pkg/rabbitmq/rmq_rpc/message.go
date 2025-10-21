package rmqrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrInvalidHeaderValue = errors.New("invalid header value")

const (
	one               = 1
	ten               = 10
	three             = 3
	internalSeparator = "internal"
	dot               = "."
	emptyString       = ""
	traceFormat       = "%s:%d %s\n"
)

type CustomError error

type MessageErrorCode int

const (
	InvalidArgument MessageErrorCode = 400
	Internal        MessageErrorCode = 500
	NotFound        MessageErrorCode = 404
	AlreadyExists   MessageErrorCode = 409
	Unauthorized    MessageErrorCode = 401
	Forbidden       MessageErrorCode = 403
)

type ErrorInfo struct {
	Reason   string            `json:"reason"`
	Metadata map[string]string `json:"metadata"`
}

type MessageError struct {
	Code       MessageErrorCode `json:"code"`
	Message    string           `json:"message" example:"message"`
	Details    ErrorInfo        `json:"details" example:"details"`
	StackTrace string           `json:"stacktrace" example:"stacktrace"`
}

func NewMessageError(code MessageErrorCode, err error) MessageError {
	sourceError := func(err error) error {
		var errs []error
		for err != nil {
			errs = append(errs, err)
			err = errors.Unwrap(err)
		}
		if len(errs) == 0 {
			return err
		}

		return errs[len(errs)-1]
	}
	return MessageError{
		Code:    code,
		Message: sourceError(err).Error(),
	}
}

func (msgErr MessageError) Error() string {
	return fmt.Sprintf("%d - %s", msgErr.Code, msgErr.Message)
}

// MessageResponse - unified message response.
type MessageResponse struct {
	Data  interface{}   `json:"data"`
	Error *MessageError `json:"error"`
}

// MessageRequest - unified message request.
type MessageRequest struct {
	Payload interface{}            `json:"payload"`
	Headers map[string]interface{} `json:"headers"`
}

func NewMessageRequest(p interface{}) *MessageRequest {
	return &MessageRequest{
		Payload: p,
		Headers: make(map[string]interface{}),
	}
}

func (msgReq *MessageRequest) AddHeader(key string, value interface{}) *MessageRequest {

	msgReq.Headers[key] = value
	return msgReq
}

func GetMesReqHeaderVal[T any](h amqp.Table, key string) (*T, error) {
	v, ok := h[key]
	if !ok {
		return nil, ErrInvalidHeaderValue
	}
	if vT, ok := v.(T); ok {
		return &vT, nil
	}
	return nil, ErrInvalidHeaderValue
}

// Pack - marshal message response.
func (r *MessageResponse) Pack() ([]byte, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc - Pack - json.Marshal: %w", err)
	}

	return body, nil
}

// Unpack - unmarshal message response.
func (r *MessageResponse) Unpack(data []byte, out interface{}) error {
	err := json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("rmq_rpc - Unpack - json.Unmarshal: %w", err)
	}

	b, err := json.Marshal(r.Data)
	if err != nil {
		return fmt.Errorf("rmq_rpc - Unpack - json.Marshal: %w", err)
	}

	err = json.Unmarshal(b, &out)
	if err != nil {
		return fmt.Errorf("rmq_rpc - Unpack - json.Unmarshal: %w", err)
	}

	return nil
}

// CheckAndCastToMessageRequest - checks and cast request to MessageRequest type.
func CheckAndCastToMessageRequest(req interface{}) (*MessageRequest, bool) {
	if msgReq, ok := req.(*MessageRequest); ok {
		return msgReq, ok
	}
	return &MessageRequest{}, false
}

// CastToMessageResponse - checks and cast response to MessageResponse type.
func CastToMessageResponse(res interface{}) *MessageResponse {
	if msg, ok := res.(*MessageResponse); ok {
		return msg
	}
	return &MessageResponse{Data: res}
}
