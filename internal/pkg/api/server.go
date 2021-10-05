package api

import (
	"context"
	"net/http"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"go.uber.org/zap"
)

type server struct {
	ctx        context.Context
	logger     *zap.Logger
	botService bots.Service
}

func New(ctx context.Context, logger *zap.Logger, botService bots.Service) *server {
	return &server{
		ctx:        ctx,
		logger:     logger,
		botService: botService,
	}
}

func (s *server) Start(addr string) error {
	s.logger.Info("starting HTTP server", zap.String("addr", addr))
	router := s.createRoutes()
	httpServer := http.Server{
		Addr:    addr,
		Handler: router,
	}

	// TODO: Shut this down somewhere else?
	go func() {
		<-s.ctx.Done()
		s.logger.Info("shutting down HTTP server")
		httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}
