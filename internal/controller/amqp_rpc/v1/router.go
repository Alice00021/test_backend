package v1

import (
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	"test_go/internal/di"
)

func NewRouter(routes map[string]server.CallHandler, uc *di.UseCase, l logger.Interface) {
	newAuthRoutes(routes, uc.Auth, l)
	newAuthorRoutes(routes, uc.Author, l)
	newBookRoutes(routes, uc.Book, l)
	newCommandRoutes(routes, uc.Command, l)
	newOperationRoutes(routes, uc.Operation, l)
}
