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

type bookRoutes struct {
	uc usecase.Book
	l  logger.Interface
}

func newBookRoutes(routes map[string]server.CallHandler, uc usecase.Book, l logger.Interface) {
	r := &bookRoutes{uc, l}
	{
		routes["v1.createBook"] = r.createBook()
		routes["v1.updateAuthor"] = r.updateBook()
		routes["v1.getBook"] = r.getBook()
		routes["v1.getAuthors"] = r.getBooks()
		routes["v1.deleteAuthor"] = r.deleteBook()
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

func (r *bookRoutes) updateBook() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var inp entity.UpdateBookInput
		if err := json.Unmarshal(d.Body, &inp); err != nil {
			r.l.Error(err, "amqp_rpc - v1 - updateBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		err := r.uc.UpdateBook(context.Background(), inp)
		if err != nil {
			if errors.Is(err, entity.ErrBookNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - updateBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}

func (r *bookRoutes) getBook() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.IdRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - V1 - getBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		res, err := r.uc.GetBook(context.Background(), req.ID)
		if err != nil {
			if errors.Is(err, entity.ErrAuthorNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - getBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *bookRoutes) getBooks() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {

		res, err := r.uc.GetBooks(context.Background())
		if err != nil {
			r.l.Error(err, "amqp_rpc - v1 - getBooks")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return res, nil
	}
}

func (r *bookRoutes) deleteBook() server.CallHandler {
	return func(d *amqp.Delivery) (interface{}, error) {
		var req request.IdRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			r.l.Error(err, "amqp_rpc - V1 - deleteBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.InvalidArgument, err)
		}

		if err := r.uc.DeleteBook(context.Background(), req.ID); err != nil {
			if errors.Is(err, entity.ErrBookNotFound) {
				return nil, rmqrpc.NewMessageError(rmqrpc.NotFound, err)
			}

			r.l.Error(err, "amqp_rpc - V1 - deleteBook")
			return nil, rmqrpc.NewMessageError(rmqrpc.Internal, err)
		}

		return nil, nil
	}
}
