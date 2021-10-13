package server

import (
	"context"
	"fmt"
	"net"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	proto2 "github.com/ch629/irc-bot-orchestrator/internal/pkg/proto"
	"github.com/ch629/irc-bot-orchestrator/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func New(logger *zap.Logger, botsService bots.Service) *server {
	return &server{
		logger:      logger,
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
		s.logger.Info("stopping gRPC server")
		grpcServer.GracefulStop()
	}()

	proto.RegisterOrchestratorServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

type server struct {
	botsService bots.Service
	logger      *zap.Logger

	proto.UnimplementedOrchestratorServer
}

// TODO: Should this be bidirectional, so the bots can send metrics back to us?
func (s *server) JoinStream(_ *proto.EmptyMessage, resp proto.Orchestrator_JoinStreamServer) error {
	// TODO: Should this ID be passed in the request instead? -> or could generate it inside of the botsService, but then have another func to notify ready?
	id := uuid.New()
	if err := resp.SendHeader(metadata.Pairs("bot_id", id.String())); err != nil {
		return fmt.Errorf("failed to set bot_id header: %w", err)
	}
	// TODO: Return a chan instead of context
	ctx := s.botsService.Join(resp.Context(), id, proto2.NewClient(resp))

	defer func() {
		s.botsService.Leave(id)
	}()

	<-ctx.Done()
	return nil
}
