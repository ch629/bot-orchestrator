package client_test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/ch629/irc-bot-orchestrator/pkg/client"
	"github.com/ch629/irc-bot-orchestrator/pkg/client/mocks"
	"github.com/ch629/irc-bot-orchestrator/pkg/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type server struct {
	proto.UnimplementedOrchestratorServer
}

func (s *server) JoinStream(_ *proto.EmptyMessage, resp proto.Orchestrator_JoinStreamServer) error {
	resp.SendHeader(metadata.Pairs("bot_id", uuid.NewString()))
	return io.EOF
}

func bufDialer(lis *bufconn.Listener) func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestJoin(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	proto.RegisterOrchestratorServer(s, &server{})
	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer(lis)), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	mockOrchestratorClient := &mocks.OrchestratorClient{}
	mockOrchestratorClient.On("JoinChannel", mock.Anything)
	mockOrchestratorClient.On("LeaveChannel", mock.Anything)

	client.Join(context.Background(), conn, mockOrchestratorClient)
}
