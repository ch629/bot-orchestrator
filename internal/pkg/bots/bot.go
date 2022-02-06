package bots

import (
	"context"
	"sync"

	"github.com/ch629/bot-orchestrator/internal/pkg/log"
	"github.com/ch629/bot-orchestrator/internal/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type bot struct {
	logger *zap.Logger
	mux    sync.Mutex
	id     uuid.UUID
	client proto.BotClient

	ctx        context.Context
	cancelFunc context.CancelFunc

	channels     map[string]struct{}
	channelSlice []string
	channelOnce  sync.Once
}

// TODO: Move this to another package?
func NewBot(ctx context.Context, id uuid.UUID, botClient proto.BotClient) Bot {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &bot{
		logger:     log.With(zap.Stringer("bot_id", id)),
		client:     botClient,
		id:         id,
		ctx:        ctx,
		cancelFunc: cancelFunc,
		channels:   make(map[string]struct{}),
	}
}

// JoinChannel notifies an individual bot to join a channel
func (b *bot) JoinChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.channels[channel] = struct{}{}
	b.invalidateChannelSlice()
	// TODO: Move logging around
	defer b.logger.Info("bot joining channel", zap.String("channel", channel))
	return b.client.SendJoinChannel(channel)
}

// LeaveChannel notifies an individual bot to leave a channel
func (b *bot) LeaveChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	if _, ok := b.channels[channel]; !ok {
		return ErrNotInChannel
	}
	delete(b.channels, channel)
	b.invalidateChannelSlice()
	// TODO: Move logging around
	defer b.logger.Info("bot leaving channel", zap.String("channel", channel))
	return b.client.SendLeaveChannel(channel)
}

// Info returns some basic information about an individual bot
func (b *bot) Info() BotInfo {
	return BotInfo{
		ID:       b.ID(),
		Channels: b.Channels(),
	}
}

// Channels returns all of the channel the bot is in
func (b *bot) Channels() []string {
	b.channelOnce.Do(func() {
		b.channelSlice = make([]string, 0, len(b.channels))
		for ch := range b.channels {
			b.channelSlice = append(b.channelSlice, ch)
		}
	})

	return b.channelSlice
}

// ID returns the ID of this bot
func (b *bot) ID() uuid.UUID {
	return b.id
}

// Remove removes the bot connection
func (b *bot) Close() {
	// TODO: Make this block until the connection is dropped?
	b.cancelFunc()
}

func (b *bot) Done() <-chan struct{} {
	return b.ctx.Done()
}

// invalidateChannelSlice invalidates the channels in the channel slice so they are regenerated
func (b *bot) invalidateChannelSlice() {
	b.channelOnce = sync.Once{}
}
