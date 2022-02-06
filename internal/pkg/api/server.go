package api

import (
	"context"
	"net/http"

	"github.com/ch629/bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/bot-orchestrator/internal/pkg/log"
	"go.uber.org/zap"
)

type server struct {
	botService bots.Service
}

func New(botService bots.Service) *server {
	return &server{
		botService: botService,
	}
}

func (s *server) Start(ctx context.Context, addr string) error {
	log.Info("starting HTTP server", zap.String("addr", addr))
	router := s.createRoutes()
	httpServer := http.Server{
		Addr:    addr,
		Handler: router,
	}

	// TODO: Shut this down somewhere else?
	go func() {
		<-ctx.Done()
		log.Info("shutting down HTTP server")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Error("failed to shutdown HTTP server", zap.Error(err))
		}
	}()

	return httpServer.ListenAndServe()
}
