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

type authRoutes struct {
	uc usecase.Auth
	l  logger.Interface
}

func newAuthRoutes(routes map[string]server.CallHandler, uc usecase.Auth, l logger.Interface) {
	r := &authRoutes{uc, l}
	{
		routes["v1.register"] = r.register()
		routes["v1.login"] = r.login()
		routes["v1.verifyEmail"] = r.verifyEmail()
		routes["v1.validationToken"] = r.validateToken()
	}
}

func (r *authRoutes) register() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp request.CreateUserRequest
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - register")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.Register(context.Background(), inp.ToEntity())
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - register")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *authRoutes) login() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp request.AuthenticateRequest
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - login")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.Login(context.Background(), inp.Username, inp.Password)
		if err != nil {
			if errors.Is(err, entity.ErrUserNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - login")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *authRoutes) verifyEmail() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp request.VerifyEmailRequest
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - verifyEmail")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		err := r.uc.VerifyEmail(context.Background(), inp.Token)
		if err != nil {
			if errors.Is(err, entity.ErrUserNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - verifyEmail")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}

func (r *authRoutes) validateToken() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.ValidateTokenRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - validateToken")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.Validation(context.Background(), req.AccessToken)
		if err != nil {
			if errors.Is(err, entity.ErrInvalidToken) || errors.Is(err, entity.ErrExpiredToken) {
				return nil, rmqrpc.NewMessageError(rmqrpc.Unauthorized, err)
			}

			r.l.Error(err, "amqp_rpc - v1 - validateToken")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}
