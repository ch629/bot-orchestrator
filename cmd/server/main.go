package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/ch629/bot-orchestrator/internal/pkg/api"
	"github.com/ch629/bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/bot-orchestrator/internal/pkg/server"
	"go.uber.org/zap"
)

// grpcurl -plaintext -import-path ./pkg/proto/ -proto orchestrator.proto -d '{}' localhost:8080 Orchestrator/JoinStream
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	botsService := bots.New(logger)
	logger.Info("starting gRPC server")
	go func() {
		if err := server.New(logger, botsService).Start(ctx, 8080); err != nil {
			logger.Fatal("failed to start gRPC server", zap.Error(err))
		}
	}()
	go func() {
		httpServer := api.New(ctx, logger, botsService)
		if err := httpServer.Start("localhost:9080"); err != nil {
			logger.Error("failed to start http server", zap.Error(err))
		}
	}()
	<-ctx.Done()
}
