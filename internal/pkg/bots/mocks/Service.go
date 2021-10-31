// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	bots "github.com/ch629/bot-orchestrator/internal/pkg/bots"

	mock "github.com/stretchr/testify/mock"

	proto "github.com/ch629/bot-orchestrator/internal/pkg/proto"

	uuid "github.com/google/uuid"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// BotInfo provides a mock function with given fields:
func (_m *Service) BotInfo() []bots.BotInfo {
	ret := _m.Called()

	var r0 []bots.BotInfo
	if rf, ok := ret.Get(0).(func() []bots.BotInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]bots.BotInfo)
		}
	}

	return r0
}

// ChannelInfo provides a mock function with given fields:
func (_m *Service) ChannelInfo() map[string][]uuid.UUID {
	ret := _m.Called()

	var r0 map[string][]uuid.UUID
	if rf, ok := ret.Get(0).(func() map[string][]uuid.UUID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]uuid.UUID)
		}
	}

	return r0
}

// DanglingChannels provides a mock function with given fields:
func (_m *Service) DanglingChannels() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Join provides a mock function with given fields: ctx, id, botClient
func (_m *Service) Join(ctx context.Context, id uuid.UUID, botClient proto.BotClient) <-chan struct{} {
	ret := _m.Called(ctx, id, botClient)

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, proto.BotClient) <-chan struct{}); ok {
		r0 = rf(ctx, id, botClient)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// JoinChannel provides a mock function with given fields: channel
func (_m *Service) JoinChannel(channel string) error {
	ret := _m.Called(channel)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(channel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Leave provides a mock function with given fields: id
func (_m *Service) Leave(id uuid.UUID) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LeaveChannel provides a mock function with given fields: channel
func (_m *Service) LeaveChannel(channel string) error {
	ret := _m.Called(channel)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(channel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveBot provides a mock function with given fields: id
func (_m *Service) RemoveBot(id uuid.UUID) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
