package v1

import (
	"test_go/internal/di"
	"test_go/pkg/logger"
	"test_go/pkg/rabbitmq/rmq_rpc/server"
)

func NewRouter(routes map[string]server.CallHandler, uc *di.UseCase, l logger.Interface) {
	newAuthorRoutes(routes, uc.Author, l)
	newBookRoutes(routes, uc.Book, l)
}
