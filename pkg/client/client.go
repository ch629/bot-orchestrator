package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ch629/bot-orchestrator/pkg/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// OrchestratorClient is a client which accepts messages from the orchestrator server
//go:generate mockery --name OrchestratorClient --disable-version-string
type OrchestratorClient interface {
	JoinChannel(channel string)
	LeaveChannel(channel string)
	Close()
}

// Join joins a bot to the orchestrator
// TODO: Call opts?
// TODO: Check that cancelling the ctx closes the bot connection properly
func Join(ctx context.Context, conn *grpc.ClientConn, client OrchestratorClient) (*uuid.UUID, error) {
	grpcClient := proto.NewOrchestratorClient(conn)
	stream, err := grpcClient.JoinStream(ctx, &proto.EmptyMessage{})
	if err != nil {
		return nil, fmt.Errorf("JoinStream: %w", err)
	}

	// Get ID from header
	var botID uuid.UUID
	{
		md, err := stream.Header()
		if err != nil {
			return nil, fmt.Errorf("Header: %w", err)
		}
		id, ok := md["bot_id"]
		if !ok || len(id) == 0 {
			return nil, errors.New("no ID provided")
		}
		if botID, err = uuid.Parse(id[0]); err != nil {
			return nil, fmt.Errorf("parse bot_id as UUID: %w", err)
		}
	}

	go func() {
		defer client.Close()
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) || resp == nil {
				return
			}

			switch resp.Type {
			case proto.StreamPayload_JOIN:
				client.JoinChannel(resp.Channel)
			case proto.StreamPayload_LEAVE:
				client.LeaveChannel(resp.Channel)
			}
		}
	}()

	return &botID, nil
}
