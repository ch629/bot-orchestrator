package server

import (
	"context"
	"fmt"
	"net"

	"github.com/ch629/bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/bot-orchestrator/internal/pkg/log"
	proto2 "github.com/ch629/bot-orchestrator/internal/pkg/proto"
	"github.com/ch629/bot-orchestrator/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func New(botsService bots.Service) *server {
	return &server{
		botsService: botsService,
	}
}

func (s *server) Start(ctx context.Context, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	go func() {
		<-ctx.Done()
		log.Info("stopping gRPC server")

		grpcServer.GracefulStop()
	}()

	proto.RegisterOrchestratorServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

type server struct {
	botsService bots.Service

	proto.UnimplementedOrchestratorServer
}

// TODO: Should this be bidirectional, so the bots can send metrics back to us?
func (s *server) JoinStream(_ *proto.EmptyMessage, resp proto.Orchestrator_JoinStreamServer) error {
	id := uuid.New()
	if err := resp.SendHeader(metadata.Pairs("bot_id", id.String())); err != nil {
		return fmt.Errorf("failed to set bot_id header: %w", err)
	}
	done := s.botsService.Join(resp.Context(), id, proto2.NewClient(resp))

	defer func() {
		if err := s.botsService.Leave(id); err != nil {
			log.Warn("failed to leave", zap.String("bot_id", id.String()), zap.Error(err))
		}
	}()

	<-done
	return nil
}
