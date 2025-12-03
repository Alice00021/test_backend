package v1

import (
	"context"
	"errors"
	"github.com/Alice00021/test_common/pkg/logger"
	rmqrpc "github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	"test_go/internal/entity"
	"test_go/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type commandRoutes struct {
	uc usecase.Command
	l  logger.Interface
}

func newCommandRoutes(routes map[string]server.CallHandler, uc usecase.Command, l logger.Interface) {
	r := &commandRoutes{uc, l}
	{
		routes["v1.updateCommands"] = r.updateCommands()
		routes["v1.getCommands"] = r.getCommands()
	}
}

func (r *commandRoutes) updateCommands() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {

		err := r.uc.UpdateCommands(context.Background())
		if err != nil {
			if errors.Is(err, entity.ErrCommandNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - updateCommands")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}

func (r *commandRoutes) getCommands() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {

		res, err := r.uc.GetCommands(context.Background())
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - getCommands")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}
