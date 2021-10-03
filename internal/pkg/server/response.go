package server

import "github.com/ch629/irc-bot-orchestrator/pkg/proto"

// Response is a wrapper around the gRPC stream to communicate with bot pods
// directly
// TODO: Rename this
type Response struct {
	stream proto.Orchestrator_JoinStreamServer
}

// SendJoinChannel sends a Join Channel request to a bot
func (r *Response) SendJoinChannel(channel string) error {
	return r.stream.Send(&proto.StreamPayload{
		Type:    proto.StreamPayload_JOIN,
		Channel: channel,
	})
}

// SendLeaveChannel sends a Leave Channel request to a bot
func (r *Response) SendLeaveChannel(channel string) error {
	return r.stream.Send(&proto.StreamPayload{
		Type:    proto.StreamPayload_LEAVE,
		Channel: channel,
	})
}
