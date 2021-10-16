package bots

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ch629/bot-orchestrator/internal/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	// ErrBotNotExist is returned when a bot is not being tracked by the orchestrator
	ErrBotNotExist = errors.New("bot does not exist")
	// ErrInChannel is returned when the orchestrator is already running in a given channel
	ErrInChannel = errors.New("already in channel")
	// ErrNotInChannel is returned when the orchestrator is not aware of a channel
	ErrNotInChannel = errors.New("not in channel")
)

//go:generate mockery --name Service --disable-version-string
type (
	// Service is an interface for bot related functionality
	Service interface {
		Join(ctx context.Context, id uuid.UUID, botClient proto.BotClient) context.Context
		Leave(id uuid.UUID) error
		RemoveBot(id uuid.UUID) error
		JoinChannel(channel string) error
		LeaveChannel(channel string) error
		BotInfo() []BotInfo
		ChannelInfo() map[string][]uuid.UUID
		DanglingChannels() []string
	}

	service struct {
		logger   *zap.Logger
		bots     map[uuid.UUID]*botState
		channels map[string][]uuid.UUID
		mux      sync.Mutex
		chanMux  sync.RWMutex
	}

	botState struct {
		logger     *zap.Logger
		mux        sync.Mutex
		id         uuid.UUID
		channels   map[string]struct{}
		client     proto.BotClient
		ctx        context.Context
		cancelFunc context.CancelFunc
	}

	// BotInfo is a struct containing basic information about a bot
	BotInfo struct {
		ID       uuid.UUID `json:"id"`
		Channels []string  `json:"channels"`
	}
)

// New creates a new service using a logger
func New(logger *zap.Logger) Service {
	return &service{
		logger:   logger,
		bots:     make(map[uuid.UUID]*botState),
		channels: make(map[string][]uuid.UUID),
	}
}

// DanglingChannels returns the channels which have no bots assigned to them
func (s *service) DanglingChannels() []string {
	channels := make([]string, 0, len(s.channels))
	for ch, bots := range s.channels {
		if len(bots) == 0 {
			channels = append(channels, ch)
		}
	}
	return channels
}

// distributeDanglingChannels distributes channels with no bots assigned to bots
func (s *service) distributeDanglingChannels() {
	danglingChannels := s.DanglingChannels()
	if len(danglingChannels) == 0 || len(s.bots) == 0 {
		return
	}
	s.logger.Info("distributing dangling channels")

	bots := s.allBots()
	for _, channel := range danglingChannels {
		// Find the bot with the least amount of channels
		sort.Sort(channelSort(bots))
		bot := bots[0]
		bot.JoinChannel(channel)
		s.channels[channel] = []uuid.UUID{bot.id}
	}
}

// Join connects a bot to the orchestrator to be controlled
func (s *service) Join(ctx context.Context, id uuid.UUID, botClient proto.BotClient) context.Context {
	logger := s.logger.With(zap.String("bot_id", id.String()))
	s.mux.Lock()
	defer s.mux.Unlock()
	ctx, cancelFunc := context.WithCancel(ctx)
	s.bots[id] = &botState{
		logger:     logger,
		client:     botClient,
		id:         id,
		ctx:        ctx,
		cancelFunc: cancelFunc,
		channels:   make(map[string]struct{}),
	}
	s.distributeDanglingChannels()
	defer logger.Info("bot joined")
	return ctx
}

// Leave removes a bot from the orchestrator
// Returns ErrBotNotExist if the bot doesn't exist
func (s *service) Leave(id uuid.UUID) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	logger := s.logger.With(zap.String("bot_id", id.String()))
	deletedBot, ok := s.bots[id]
	if !ok {
		return ErrBotNotExist
	}
	delete(s.bots, id)
	// Delete channel references to this bot
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	for ch := range deletedBot.channels {
		newChannelIDs := make([]uuid.UUID, 0, len(s.channels[ch])-1)
		for _, old := range s.channels[ch] {
			if old != id {
				newChannelIDs = append(newChannelIDs, old)
			}
		}
		s.channels[ch] = newChannelIDs
	}
	s.distributeDanglingChannels()
	logger.Info("bot left")
	return nil
}

// RemoveBot removes a bot from the orchestrator
func (s *service) RemoveBot(id uuid.UUID) error {
	bot, ok := s.bots[id]
	if !ok {
		return ErrBotNotExist
	}
	bot.logger.Info("removing bot")
	bot.cancelFunc()
	return nil
}

// JoinChannel notifies a bot to connect to a channel, assigned to the bot with the least current channels
// Returns ErrInChannel if the orchestrator is already in the channel
// TODO: Come up with a better way to weigh the bots based on message throughput vs just channel count
// TODO: How do we get this to join on multiple bots to keep high availability?
func (s *service) JoinChannel(channel string) error {
	s.chanMux.RLock()
	_, ok := s.channels[channel]
	s.chanMux.RUnlock()
	if ok {
		return ErrInChannel
	}

	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	if len(s.bots) == 0 {
		s.channels[channel] = make([]uuid.UUID, 0)
		return nil
	}

	bots := s.allBots()
	// Find the bot with the least amount of channels
	sort.Sort(channelSort(bots))
	bot := bots[0]
	bot.JoinChannel(channel)
	s.channels[channel] = []uuid.UUID{bot.id}
	return nil
}

func (s *service) allBots() []*botState {
	bots := make([]*botState, len(s.bots))
	i := 0
	for _, bot := range s.bots {
		bots[i] = bot
		i++
	}
	return bots
}

// LeaveChannel notifies any bot connected to a channel to leave & stops tracking it
// Returns ErrNotInChannel if the bot isn't in the given channel
func (s *service) LeaveChannel(channel string) error {
	s.chanMux.RLock()
	botIds, ok := s.channels[channel]
	s.chanMux.RUnlock()
	if !ok {
		return ErrNotInChannel
	}
	s.chanMux.Lock()
	defer s.chanMux.Unlock()

	var err error
	for _, id := range botIds {
		bot := s.bots[id]
		if leaveErr := bot.LeaveChannel(channel); err != nil {
			err = multierr.Append(err, fmt.Errorf("%s: %w", id, leaveErr))
		}
	}
	delete(s.channels, channel)

	return err
}

// BotInfo returns some basic info about all connected bots
func (s *service) BotInfo() []BotInfo {
	botInfos := make([]BotInfo, 0, len(s.bots))
	for _, bot := range s.bots {
		botInfos = append(botInfos, bot.BotInfo())
	}
	return botInfos
}

// JoinChannel notifies an individual bot to join a channel
func (b *botState) JoinChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.channels[channel] = struct{}{}
	// TODO: Move logging around
	defer b.logger.Info("bot joining channel", zap.String("channel", channel))
	return b.client.SendJoinChannel(channel)
}

// LeaveChannel notifies an individual bot to leave a channel
func (b *botState) LeaveChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	if _, ok := b.channels[channel]; !ok {
		return ErrNotInChannel
	}
	delete(b.channels, channel)
	// TODO: Move logging around
	defer b.logger.Info("bot leaving channel", zap.String("channel", channel))
	return b.client.SendLeaveChannel(channel)
}

// BotInfo returns some basic information about an individual bot
func (b *botState) BotInfo() BotInfo {
	channels := make([]string, 0, len(b.channels))
	for ch := range b.channels {
		channels = append(channels, ch)
	}
	return BotInfo{
		ID:       b.id,
		Channels: channels,
	}
}

// ChannelInfo returns information about which bots are connected to each channel
func (s *service) ChannelInfo() map[string][]uuid.UUID {
	s.chanMux.RLock()
	defer s.chanMux.RUnlock()
	return s.channels
}
