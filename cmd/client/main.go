package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/ch629/irc-bot-orchestrator/pkg/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(log)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	conn, err := grpc.DialContext(ctx, ":8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("failed to dial grpc", zap.Error(err))
	}
	id, err := client.Join(ctx, conn, newLogClient(log, cancel))
	if err != nil {
		log.Fatal("failed to join client", zap.Error(err))
	}
	log.Info("joined orchestrator", zap.String("bot_id", id.String()))
	<-ctx.Done()
}

func newLogClient(logger *zap.Logger, cancel context.CancelFunc) *LogClient {
	return &LogClient{
		logger: logger,
		cancel: cancel,
	}
}

type LogClient struct {
	logger *zap.Logger
	cancel context.CancelFunc
}

func (c *LogClient) JoinChannel(channel string) {
	c.logger.Info("joining", zap.String("channel", channel))
}

func (c *LogClient) LeaveChannel(channel string) {
	c.logger.Info("leaving", zap.String("channel", channel))
}

func (c *LogClient) Close() {
	c.logger.Info("close")
	c.cancel()
}
