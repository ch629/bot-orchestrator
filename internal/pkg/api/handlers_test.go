package api

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_ServerJoinChannel(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(mockBotService *mocks.Service)
		payload    string
		assertions func(t *testing.T, resp http.Response)
	}{
		{
			name: "Success: Valid request",
			setupMocks: func(mockBotService *mocks.Service) {
				mockBotService.On("JoinChannel", "foo").Return(nil)
			},
			payload: `{"channel": "foo"}`,
			assertions: func(t *testing.T, resp http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name:    "Failure: Invalid JSON request",
			payload: `{`,
			assertions: func(t *testing.T, resp http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				bs, _ := ioutil.ReadAll(resp.Body)
				assert.JSONEq(t, `{"error":"received json invalid request body: unexpected EOF"}`, string(bs))
			},
		},
		{
			name:    "Failure: No channel given in JSON body",
			payload: `{}`,
			assertions: func(t *testing.T, resp http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				bs, _ := ioutil.ReadAll(resp.Body)
				assert.JSONEq(t, `{"error":"missing channel in request"}`, string(bs))
			},
		},
		{
			name: "Failure: Error joining channel",
			setupMocks: func(mockBotService *mocks.Service) {
				mockBotService.On("JoinChannel", "foo").Return(errors.New("failure"))
			},
			payload: `{"channel": "foo"}`,
			assertions: func(t *testing.T, resp http.Response) {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				bs, _ := ioutil.ReadAll(resp.Body)
				assert.JSONEq(t, `{"error":"failed to join channel: failure"}`, string(bs))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/join", strings.NewReader(tt.payload))
			rw := httptest.NewRecorder()
			mockBotsService := &mocks.Service{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockBotsService)
			}
			server := New(context.Background(), zaptest.NewLogger(t), mockBotsService)
			server.JoinChannel().ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()
			tt.assertions(t, *res)
			mockBotsService.AssertExpectations(t)
		})
	}
}
