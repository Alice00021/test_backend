package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"test_go/internal/di"
	"test_go/pkg/jwt"
	"test_go/pkg/rabbitmq/rmq_rpc/client"
	"test_go/pkg/rabbitmq/rmq_rpc/server"
	"test_go/pkg/transactional"

	"test_go/pkg/httpserver"
	"test_go/pkg/logger"
	"test_go/pkg/postgres"

	"github.com/gin-gonic/gin"

	"test_go/config"

	amqprpc "test_go/internal/controller/amqp_rpc"
	"test_go/internal/controller/http"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	jwtManager := jwt.NewJWTManager(cfg.JWT.SecretKey)

	// Transaction builder
	pgTx := transactional.NewPgTransaction(pg)

	// RabbitMQ RPC Client
	rmqClient, err := client.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, cfg.RMQ.ClientExchange, cfg.App.Name, cfg.RMQ.ClientPrefix)
	if err != nil {
		l.Fatal("RabbitMQ RPC Client - init error - client.New")
	}

	// Repo
	repo := di.NewRepo(pg)

	// Use-Case
	uc := di.NewUseCase(pgTx, repo, l, cfg, jwtManager)

	// RabbitMQ RPC Server
	rmqRouter := amqprpc.NewRouter(uc, l)

	rmqServer, err := server.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, cfg.App.Name, rmqRouter, l, cfg.RMQ.ClientPrefix)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - rmqServer - server.New: %w", err))
	}

	// HTTP Server
	handler := gin.New()
	http.NewRouter(handler, cfg, l, uc)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	case err = <-rmqServer.Notify():
		l.Error(fmt.Errorf("app - Run - rmqServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	err = rmqServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - rmqServer.Shutdown: %w", err))
	}

	err = rmqClient.Shutdown()
	if err != nil {
		l.Fatal("RabbitMQ RPC Client - shutdown error - rmqClient.RemoteCall", err)
	}
}
