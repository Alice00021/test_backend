package v1

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alice00021/test_common/pkg/logger"
	rmqrpc "github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	amqp "github.com/rabbitmq/amqp091-go"
	"test_go/internal/controller/amqp_rpc/v1/request"
	"test_go/internal/entity"
	"test_go/internal/usecase"
)

type operationRoutes struct {
	uc usecase.Operation
	l  logger.Interface
}

func newOperationRoutes(routes map[string]server.CallHandler, uc usecase.Operation, l logger.Interface) {
	r := &operationRoutes{uc, l}
	{
		routes["v1.createOperation"] = r.createOperation()
		routes["v1.updateOperation"] = r.updateOperation()
		routes["v1.getOperation"] = r.getOperation()
		routes["v1.getOperations"] = r.getOperations()
		routes["v1.deleteOperation"] = r.deleteOperation()
	}
}

func (r *operationRoutes) createOperation() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp entity.CreateOperationInput
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - createOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.CreateOperation(context.Background(), inp)
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - createOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *operationRoutes) updateOperation() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp entity.UpdateOperationInput
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - updateOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		err := r.uc.UpdateOperation(context.Background(), inp)
		if err != nil {
			if errors.Is(err, entity.ErrOperationNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - updateOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}

func (r *operationRoutes) getOperation() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.IdRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - V1 - getOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.GetOperation(context.Background(), req.ID)
		if err != nil {
			if errors.Is(err, entity.ErrOperationNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - getOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *operationRoutes) getOperations() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {

		res, err := r.uc.GetOperations(context.Background())
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - getOperations")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *operationRoutes) deleteOperation() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.IdRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - V1 - deleteOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		if err := r.uc.DeleteOperation(context.Background(), req.ID); err != nil {
			if errors.Is(err, entity.ErrOperationNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - deleteOperation")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}
