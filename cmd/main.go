package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/server"
	"go.uber.org/zap"
)

// grpcurl -plaintext -import-path . -proto orchestrator.proto -d '{}' localhost:8080 Orchestrator/Join
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	botsService := bots.New(logger)
	logger.Info("starting gRPC server")
	if err := server.New(logger, botsService).Start(ctx, 8080); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}
}
