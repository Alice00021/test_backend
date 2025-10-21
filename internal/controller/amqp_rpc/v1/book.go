package v1

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"test_go/internal/entity"
	"test_go/internal/usecase"
	"test_go/pkg/logger"
	rmqrpc "test_go/pkg/rabbitmq/rmq_rpc"
	"test_go/pkg/rabbitmq/rmq_rpc/server"
)

type bookRoutes struct {
	uc usecase.Book
	l  logger.Interface
}

func newBookRoutes(routes map[string]server.CallHandler, uc usecase.Book, l logger.Interface) {
	r := &bookRoutes{uc, l}
	{
		routes["v1.createBook"] = r.createBook()

	}
}

func (r *bookRoutes) createBook() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp entity.CreateBookInput
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - createBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.CreateBook(context.Background(), inp)
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - createBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}
