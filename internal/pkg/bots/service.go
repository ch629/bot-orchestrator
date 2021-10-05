package bots

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/proto"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	ErrNoBots       = errors.New("no bots")
	ErrBotNotExist  = errors.New("bot does not exist")
	ErrInChannel    = errors.New("already in channel")
	ErrNotInChannel = errors.New("not in channel")
)

//go:generate mockery --name Service --disable-version-string
type Service interface {
	Join(ctx context.Context, botClient proto.BotClient) (context.Context, uuid.UUID)
	Leave(id uuid.UUID)
	RemoveBot(id uuid.UUID) error
	JoinChannel(channel string) error
	LeaveChannel(channel string) error
	BotInfo() []BotInfo
}

// TODO: RWMutex on channels?
type service struct {
	logger   *zap.Logger
	bots     map[uuid.UUID]*botState
	channels map[string][]uuid.UUID
	mux      sync.Mutex
}

type botState struct {
	logger     *zap.Logger
	mux        sync.Mutex
	id         uuid.UUID
	channels   map[string]struct{}
	client     proto.BotClient
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type BotInfo struct {
	ID       uuid.UUID `json:"id"`
	Channels []string  `json:"channels"`
}

func New(logger *zap.Logger) Service {
	return &service{
		logger:   logger,
		bots:     make(map[uuid.UUID]*botState),
		channels: make(map[string][]uuid.UUID),
	}
}

// TODO: How do we make this horizontally scalable?
// TODO: Rebalancing
func (s *service) Join(ctx context.Context, botClient proto.BotClient) (context.Context, uuid.UUID) {
	id := uuid.New()
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
	defer logger.Info("bot joined")
	return ctx, id
}

// TODO: Need to rebalance channels to another bot, or remove from map if no bots remain
func (s *service) Leave(id uuid.UUID) {
	s.mux.Lock()
	defer s.mux.Unlock()
	logger := s.logger.With(zap.String("bot_id", id.String()))
	delete(s.bots, id)
	logger.Info("bot left")
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

// TODO: Come up with a better way to weigh the bots based on message throughput vs just channel count
// TODO: How do we get this to join on multiple bots to keep high availability?
func (s *service) JoinChannel(channel string) error {
	if len(s.bots) == 0 {
		return ErrNoBots
	}

	if _, ok := s.channels[channel]; ok {
		return ErrInChannel
	}

	bots := make([]*botState, len(s.bots))
	i := 0
	for _, bot := range s.bots {
		bots[i] = bot
		i++
	}
	// Find the bot with the least amount of channels
	sort.Sort(channelSort(bots))
	bot := bots[0]
	bot.JoinChannel(channel)
	s.channels[channel] = []uuid.UUID{bot.id}
	return nil
}

func (s *service) LeaveChannel(channel string) error {
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

func (s *service) BotInfo() []BotInfo {
	botInfos := make([]BotInfo, 0, len(s.bots))
	for _, bot := range s.bots {
		botInfos = append(botInfos, bot.BotInfo())
	}
	return botInfos
}

func (b *botState) JoinChannel(channel string) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.channels[channel] = struct{}{}
	// TODO: Move logging around
	defer b.logger.Info("bot joining channel", zap.String("channel", channel))
	return b.client.SendJoinChannel(channel)
}

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
