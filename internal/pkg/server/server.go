package server

import (
	"context"
	"fmt"
	"net"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/domain"
	"github.com/ch629/irc-bot-orchestrator/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func New(botsService bots.Service) *server {
	return &server{
		botsService: botsService,
	}
}

func (s *server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterOrchestratorServer(grpcServer, s)
	// TODO: grpcServer.GracefulStop()
	return grpcServer.Serve(lis)
}

type server struct {
	botsService bots.Service

	proto.UnimplementedOrchestratorServer
}

func (s *server) Join(ctx context.Context, req *proto.BotJoinRequest) (*proto.BotJoinResponse, error) {
	return &proto.BotJoinResponse{
		Id: "abc",
	}, nil
}

// TODO: a permanent gRPC stream from bots to this orchestrator seems like it's not the best approach, but I can't think of an alternative right now
// TODO: Should this be bidirectional, so the bots can send metrics back to us?
func (s *server) JoinStream(req *proto.BotJoinResponse, resp proto.Orchestrator_JoinStreamServer) error {
	id, ch := s.botsService.Join()
	resp.SendHeader(metadata.Pairs("bot_id", id.String()))

	defer func() {
		s.botsService.Leave(id)
	}()

loop:
	for {
		select {
		case <-resp.Context().Done():
			break loop
		case msg, ok := <-ch:
			if !ok {
				break loop
			}
			switch m := msg.(type) {
			case domain.JoinRequest:
				resp.Send(MapJoinToProto(m))
			case domain.LeaveRequest:
				resp.Send(MapLeaveToProto(m))
			}
		}
	}
	fmt.Println("Exiting")
	return nil
}
