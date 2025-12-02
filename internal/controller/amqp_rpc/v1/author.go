package v1

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alice00021/test_common/pkg/logger"
	rmqrpc "github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	"test_go/internal/controller/amqp_rpc/v1/request"
	"test_go/internal/entity"
	"test_go/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type authorRoutes struct {
	uc usecase.Author
	l  logger.Interface
}

func newAuthorRoutes(routes map[string]server.CallHandler, uc usecase.Author, l logger.Interface) {
	r := &authorRoutes{uc, l}
	{
		routes["v1.getAuthor"] = r.getAuthor()
	}
}

func (r *authorRoutes) getAuthor() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.IdRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - V1 - getAuthor")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.GetAuthor(context.Background(), req.ID)
		if err != nil {
			if errors.Is(err, entity.ErrAuthorNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - getAuthor")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}
