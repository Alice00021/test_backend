package v1

import (
	v1 "test_go/internal/controller/amqp_rpc/v1"
	"test_go/internal/di"
	"test_go/pkg/logger"
	"test_go/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(uc *di.UseCase, l logger.Interface) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		v1.NewRouter(routes, uc, l)
	}

	return routes
}
