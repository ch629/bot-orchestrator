package bots

import "github.com/google/uuid"

type (
	Bot interface {
		ID() uuid.UUID
		Channels() []string
		JoinChannel(channel string) error
		LeaveChannel(channel string) error
		Close()
		Info() BotInfo
		Done() <-chan struct{}
	}

	// BotInfo is a struct containing basic information about a bot
	// TODO: JSON tags or map to a DTO?
	BotInfo struct {
		ID       uuid.UUID `json:"id"`
		Channels []string  `json:"channels"`
	}
)
