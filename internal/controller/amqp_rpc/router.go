package v1

import (
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	v1 "test_go/internal/controller/amqp_rpc/v1"
	"test_go/internal/di"
)

// NewRouter -.
func NewRouter(uc *di.UseCase, l logger.Interface) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		v1.NewRouter(routes, uc, l)
	}

	return routes
}
