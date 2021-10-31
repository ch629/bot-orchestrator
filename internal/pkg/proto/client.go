package proto

import "github.com/ch629/bot-orchestrator/pkg/proto"

// BotClient is a client to send messages to an individual bot
//go:generate mockery --name BotClient --disable-version-string
type BotClient interface {
	SendJoinChannel(channel string) error
	SendLeaveChannel(channel string) error
}

// NewClient builds a new BotClient using a gRPC stream
func NewClient(stream proto.Orchestrator_JoinStreamServer) BotClient {
	return &botClient{
		stream: stream,
	}
}

// BotClient is a wrapper around the gRPC stream to communicate with bot pods
// directly
type botClient struct {
	stream proto.Orchestrator_JoinStreamServer
}

// SendJoinChannel sends a Join Channel request to a bot
func (c *botClient) SendJoinChannel(channel string) error {
	return c.stream.Send(&proto.StreamPayload{
		Type:    proto.StreamPayload_JOIN,
		Channel: channel,
	})
}

// SendLeaveChannel sends a Leave Channel request to a bot
func (c *botClient) SendLeaveChannel(channel string) error {
	return c.stream.Send(&proto.StreamPayload{
		Type:    proto.StreamPayload_LEAVE,
		Channel: channel,
	})
}
