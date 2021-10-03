package server

import (
	"context"
	"fmt"
	"net"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/irc-bot-orchestrator/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func New(logger *zap.Logger, botsService *bots.Service) *server {
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
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	go func() {
		<-ctx.Done()
		s.logger.Info("stopping gRPC server")
		grpcServer.GracefulStop()
	}()

	proto.RegisterOrchestratorServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

type server struct {
	botsService *bots.Service
	logger      *zap.Logger

	proto.UnimplementedOrchestratorServer
}

// TODO: a permanent gRPC stream from bots to this orchestrator seems like it's not the best approach, but I can't think of an alternative right now
// TODO: Should this be bidirectional, so the bots can send metrics back to us?
func (s *server) JoinStream(_ *proto.EmptyMessage, resp proto.Orchestrator_JoinStreamServer) error {
	id := s.botsService.Join(Response{resp})
	resp.SendHeader(metadata.Pairs("bot_id", id.String()))

	defer func() {
		s.botsService.Leave(id)
	}()

	<-resp.Context().Done()
	s.logger.Info("exiting")
	return nil
}
