package app

import (
	"fmt"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/client"
	"github.com/Alice00021/test_common/pkg/rabbitmq/rmq_rpc/server"
	"github.com/Alice00021/test_common/pkg/transactional"
	"os"
	"os/signal"
	"syscall"
	"test_go/internal/di"

	"github.com/Alice00021/test_common/pkg/httpserver"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/mongodb"
	"github.com/Alice00021/test_common/pkg/postgres"

	"github.com/gin-gonic/gin"

	"test_go/config"

	amqprpc "test_go/internal/controller/amqp_rpc"
	"test_go/internal/controller/http"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.NewMultipleWriter(
		logger.Level(cfg.Log.Level),
		logger.FileName(cfg.Log.FileName),
	)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// MongoDB Client
	mongoClient, err := mongodb.New(mongodb.Config{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
		Timeout:  cfg.MongoDB.Timeout,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("failed to init mongodb: %w", err))
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			l.Error(fmt.Errorf("mongodb close error: %w", err))
		}
	}()

	// Transaction builder
	pgTx := transactional.NewPgTransaction(pg)

	// RabbitMQ RPC Client
	rmqClient, err := client.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, cfg.RMQ.ClientExchange, cfg.App.Name, cfg.RMQ.ClientPrefix)
	if err != nil {
		l.Fatal("RabbitMQ RPC Client - init error - client.New")
	}

	// Repo
	repo := di.NewRepo(pg, mongoClient)

	// Use-Case
	uc := di.NewUseCase(pgTx, repo, l, cfg)

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
