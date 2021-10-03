package bots

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/server"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	ErrNoBots       = errors.New("no bots")
	ErrInChannel    = errors.New("already in channel")
	ErrNotInChannel = errors.New("not in channel")
)

// TODO: RWMutex on channels?
type Service struct {
	logger   *zap.Logger
	bots     map[uuid.UUID]*BotState
	channels map[string][]uuid.UUID
	mux      sync.Mutex
}

type BotState struct {
	logger   *zap.Logger
	mux      sync.Mutex
	id       uuid.UUID
	channels map[string]struct{}
	response server.Response
}

func New(logger *zap.Logger) *Service {
	return &Service{
		logger:   logger,
		bots:     make(map[uuid.UUID]*BotState),
		channels: make(map[string][]uuid.UUID),
	}
}

// TODO: Replace interface{}
// TODO: How do we make this horizontally scalable?
// TODO: Rebalancing
func (s *Service) Join(response server.Response) uuid.UUID {
	id := uuid.New()
	logger := s.logger.With(zap.String("bot_id", id.String()))
	s.mux.Lock()
	defer s.mux.Unlock()
	s.bots[id] = &BotState{
		logger:   logger,
		response: response,
		id:       id,
	}
	defer logger.Info("bot joined")
	return id
}

func (s *Service) Leave(id uuid.UUID) {
	s.mux.Lock()
	defer s.mux.Unlock()
	logger := s.logger.With(zap.String("bot_id", id.String()))
	delete(s.bots, id)
	logger.Info("bot left")
}

// TODO: Come up with a better way to weigh the bots based on message throughput vs just channel count
// TODO: How do we get this to join on multiple bots to keep high availability?
func (s *Service) JoinChannel(channel string) error {
	if len(s.bots) == 0 {
		return ErrNoBots
	}

	if _, ok := s.channels[channel]; ok {
		return ErrInChannel
	}

	bots := make([]*BotState, len(s.bots))
	i := 0
	for _, bot := range bots {
		bots[i] = bot
		i++
	}
	// Find the bot with the least amount of channels
	sort.Sort(channelSort(bots))
	bot := bots[len(bots)-1]
	bot.JoinChannel(channel)
	s.channels[channel] = []uuid.UUID{bot.id}
	return nil
}

func (s *Service) LeaveChannel(channel string) error {
	s.mux.Lock()
	// TODO: Can we unlock this earlier?
	defer s.mux.Unlock()
	botIds, ok := s.channels[channel]
	if !ok {
		return ErrNotInChannel
	}

	var err error
	for _, id := range botIds {
		bot := s.bots[id]
		if leaveErr := bot.LeaveChannel(channel); err != nil {
			err = multierr.Append(err, fmt.Errorf("%s: %w", id, leaveErr))
		}
	}

	return err
}

func (b *BotState) JoinChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.channels[channel] = struct{}{}
	// TODO: Move logging around
	defer b.logger.Info("bot joining channel", zap.String("channel", channel))
	return b.response.SendJoinChannel(channel)
}

func (b *BotState) LeaveChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	if _, ok := b.channels[channel]; !ok {
		return ErrNotInChannel
	}
	delete(b.channels, channel)
	// TODO: Move logging around
	defer b.logger.Info("bot leaving channel", zap.String("channel", channel))
	return b.response.SendLeaveChannel(channel)
}
