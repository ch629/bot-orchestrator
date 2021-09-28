package server

import (
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/domain"
	"github.com/ch629/irc-bot-orchestrator/pkg/proto"
)

func MapJoinToProto(join domain.JoinRequest) *proto.StreamPayload {
	return &proto.StreamPayload{
		Type:    proto.StreamPayload_JOIN,
		Channel: join.Channel,
	}
}

func MapLeaveToProto(leave domain.LeaveRequest) *proto.StreamPayload {
	return &proto.StreamPayload{
		Type:    proto.StreamPayload_LEAVE,
		Channel: leave.Channel,
	}
}
