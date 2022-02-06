package bots

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ch629/bot-orchestrator/internal/pkg/log"
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
		Join(ctx context.Context, id uuid.UUID, botClient proto.BotClient) <-chan struct{}
		Leave(id uuid.UUID) error
		RemoveBot(id uuid.UUID) error
		JoinChannel(channel string) error
		LeaveChannel(channel string) error
		BotInfo() []BotInfo
		ChannelInfo() map[string][]uuid.UUID
		DanglingChannels() []string
	}

	service struct {
		bots   map[uuid.UUID]Bot
		botMux sync.Mutex

		channels map[string][]uuid.UUID
		chanMux  sync.RWMutex
	}
)

// New creates a new service using a logger
func New(logger *zap.Logger) Service {
	return &service{
		bots:     make(map[uuid.UUID]Bot),
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
	log.Info("distributing dangling channels")

	bots := s.allBots()
	for _, channel := range danglingChannels {
		// Find the bot with the least amount of channels
		sort.Sort(sortByChannelLen(bots))
		bot := bots[0]
		if err := bot.JoinChannel(channel); err != nil {
			log.Warn("failed to join channel",
				zap.String("channel", channel),
				zap.Stringer("bot_id", bot.ID()),
				zap.Error(err),
			)
			continue
		}
		s.channels[channel] = []uuid.UUID{bot.ID()}
	}
}

// Join connects a bot to the orchestrator to be controlled
func (s *service) Join(ctx context.Context, id uuid.UUID, botClient proto.BotClient) <-chan struct{} {
	s.botMux.Lock()
	defer s.botMux.Unlock()
	bot := NewBot(ctx, id, botClient)
	s.bots[id] = bot
	// TODO: Should this be async? -> Breaks tests if it is
	s.distributeDanglingChannels()
	log.Info("bot joined", zap.Stringer("bot_id", id))
	return bot.Done()
}

// Leave removes a bot from the orchestrator
// Returns ErrBotNotExist if the bot doesn't exist
func (s *service) Leave(id uuid.UUID) error {
	s.botMux.Lock()
	defer s.botMux.Unlock()
	deletedBot, ok := s.bots[id]
	if !ok {
		return ErrBotNotExist
	}
	delete(s.bots, id)
	// Delete channel references to this bot
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	for _, ch := range deletedBot.Channels() {
		newChannelIDs := make([]uuid.UUID, 0, len(s.channels[ch])-1)
		for _, old := range s.channels[ch] {
			if old != id {
				newChannelIDs = append(newChannelIDs, old)
			}
		}
		s.channels[ch] = newChannelIDs
	}
	s.distributeDanglingChannels()
	log.Info("bot left", zap.Stringer("bot_id", id))
	return nil
}

// RemoveBot removes a bot from the orchestrator
func (s *service) RemoveBot(id uuid.UUID) error {
	bot, ok := s.bots[id]
	if !ok {
		return ErrBotNotExist
	}
	bot.Close()
	return nil
}

// JoinChannel notifies a bot to connect to a channel, assigned to the bot with the least current channels
// Returns ErrInChannel if the orchestrator is already in the channel
// TODO: Come up with a better way to weigh the bots based on message throughput vs just channel count
// TODO: How do we get this to join on multiple bots to keep high availability?
func (s *service) JoinChannel(channel string) error {
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	if _, ok := s.channels[channel]; ok {
		return ErrInChannel
	}

	if len(s.bots) == 0 {
		s.channels[channel] = make([]uuid.UUID, 0)
		return nil
	}

	bots := s.allBots()
	// Find the bot with the least amount of channels
	sort.Sort(sortByChannelLen(bots))
	bot := bots[0]
	if err := bot.JoinChannel(channel); err != nil {
		return fmt.Errorf("bot.JoinChannel: %w", err)
	}
	s.channels[channel] = []uuid.UUID{bot.ID()}
	return nil
}

func (s *service) allBots() []Bot {
	bots := make([]Bot, 0, len(s.bots))
	for _, bot := range s.bots {
		bots = append(bots, bot)
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
		botInfos = append(botInfos, bot.Info())
	}
	return botInfos
}

// ChannelInfo returns information about which bots are connected to each channel
func (s *service) ChannelInfo() map[string][]uuid.UUID {
	s.chanMux.RLock()
	defer s.chanMux.RUnlock()
	return s.channels
}
