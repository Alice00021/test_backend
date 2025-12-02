package v1

import (
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	"test_go/internal/di"
)

func NewRouter(routes map[string]server.CallHandler, uc *di.UseCase, l logger.Interface) {
	newAuthorRoutes(routes, uc.Author, l)
	newBookRoutes(routes, uc.Book, l)
}
