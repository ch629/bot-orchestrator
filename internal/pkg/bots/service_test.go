package bots_test

import (
	"context"
	"testing"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/proto/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_ServiceJoin(t *testing.T) {
	service := bots.New(zap.NewNop())
	mockBotClient := &mocks.BotClient{}
	mockBotClient.On("SendJoinChannel", "foo").Return(nil)
	service.Join(context.Background(), mockBotClient)
	service.JoinChannel("foo")
	botInfo := service.BotInfo()
	require.Len(t, botInfo, 1)
	require.Equal(t, []string{
		"foo",
	}, botInfo[0].Channels)
	require.Contains(t, service.ChannelInfo(), "foo")
	mockBotClient.AssertExpectations(t)
}

func Test_ServiceDanglingChannels(t *testing.T) {
	service := bots.New(zap.NewNop())
	mockBotClient := &mocks.BotClient{}
	mockBotClient.On("SendJoinChannel", "foo").Return(nil)
	_, id := service.Join(context.Background(), mockBotClient)
	service.JoinChannel("foo")

	// Bot leaves, so we have a channel with no bots assigned
	service.Leave(id)
	require.Equal(t, []string{"foo"}, service.DanglingChannels(), "there should be one dangling channel left")
	require.Contains(t, service.ChannelInfo(), "foo")
	// New bot joins, so dangling channels should be assigned
	service.Join(context.Background(), mockBotClient)
	require.Empty(t, service.DanglingChannels(), "all dangling channels should be assigned to the new bot")
	require.Contains(t, service.ChannelInfo(), "foo")
	mockBotClient.AssertExpectations(t)
}

func Test_ServiceJoinChannel(t *testing.T) {
	mockBotClient := &mocks.BotClient{}
	mockBotClient.On("SendJoinChannel", "foo").Return(nil)
	service := bots.New(zap.NewNop())
	// First join should be successful
	require.NoError(t, service.JoinChannel("foo"))
	// We're already in the channel
	require.ErrorIs(t, service.JoinChannel("foo"), bots.ErrInChannel)
	require.Contains(t, service.ChannelInfo(), "foo")
	require.Len(t, service.DanglingChannels(), 1)
	// All dangling channels should be assigned to the new bot
	service.Join(context.Background(), mockBotClient)
	require.Len(t, service.DanglingChannels(), 0, "all dangling channels should now be assigned")
	require.Contains(t, service.BotInfo()[0].Channels, "foo")
}

func Test_ServiceLeaveChannel(t *testing.T) {
	service := bots.New(zap.NewNop())
	mockBotClient := &mocks.BotClient{}
	mockBotClient.On("SendJoinChannel", "foo").Return(nil)
	mockBotClient.On("SendLeaveChannel", "foo").Return(nil)
	service.Join(context.Background(), mockBotClient)
	service.JoinChannel("foo")
	botInfo := service.BotInfo()
	require.Len(t, botInfo, 1)
	require.Contains(t, botInfo[0].Channels, "foo")
	service.LeaveChannel("foo")
	botInfo = service.BotInfo()
	require.NotContains(t, botInfo[0].Channels, "foo")
	require.NotContains(t, service.ChannelInfo(), "foo")
	mockBotClient.AssertExpectations(t)
}

func Test_ServiceLeaveMultiple(t *testing.T) {
	service := bots.New(zap.NewNop())
	mockBotClient := &mocks.BotClient{}
	mockBotClient.On("SendJoinChannel", mock.Anything).Return(nil)
	mockBotClient.On("SendLeaveChannel", mock.Anything).Return(nil)
	_, id1 := service.Join(context.Background(), mockBotClient)
	_, _ = service.Join(context.Background(), mockBotClient)
	service.JoinChannel("foo")
	service.JoinChannel("bar")
	service.JoinChannel("baz")
	service.Leave(id1)
	botInfo := service.BotInfo()
	require.Len(t, botInfo, 1)
	require.Len(t, botInfo[0].Channels, 3)
	require.Contains(t, botInfo[0].Channels, "foo")
	require.Contains(t, botInfo[0].Channels, "bar")
	require.Contains(t, botInfo[0].Channels, "baz")
	require.Empty(t, service.DanglingChannels())
}
