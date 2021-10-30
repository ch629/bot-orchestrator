package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ch629/bot-orchestrator/pkg/client"
	"github.com/ch629/bot-orchestrator/pkg/client/mocks"
	"github.com/ch629/bot-orchestrator/pkg/proto"
	"github.com/google/uuid"
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
	return nil
}

func bufDialer(lis *bufconn.Listener) func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return lis.Dial()
	}
}

// TODO: Test receiving join & leave messages
func TestJoin(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	proto.RegisterOrchestratorServer(s, &server{})
	go s.Serve(lis)
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer(lis)), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	mockOrchestratorClient := &mocks.OrchestratorClient{}
	mockOrchestratorClient.On("Close")

	ctx, cancel := context.WithCancel(context.Background())
	_, err = client.Join(ctx, conn, mockOrchestratorClient)
	require.NoError(t, err)
	cancel()
	lis.Close()
	// Have to wait until the goroutine closes
	time.Sleep(10 * time.Millisecond)
	mockOrchestratorClient.AssertExpectations(t)
}
