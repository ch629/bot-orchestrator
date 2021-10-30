package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// JoinChannel is the handler to tell a bot to join a channel
func (s *server) JoinChannel() http.HandlerFunc {
	type request struct {
		Channel string `json:"channel"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// TODO: Handle
			_ = writeErr(rw, fmt.Errorf("received json invalid request body: %w", err), http.StatusBadRequest)
			return
		}

		if req.Channel == "" {
			// TODO: Handle
			_ = writeErr(rw, errors.New("missing channel in request"), http.StatusBadRequest)
			return
		}

		if err := s.botService.JoinChannel(req.Channel); err != nil {
			// TODO: Handle
			_ = writeErr(rw, fmt.Errorf("failed to join channel: %w", err), http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (s *server) LeaveChannel() http.HandlerFunc {
	type request struct {
		Channel string `json:"channel"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// TODO: Handle
			_ = writeErr(rw, fmt.Errorf("received json invalid request body: %w", err), http.StatusBadRequest)
			return
		}

		if req.Channel == "" {
			// TODO: Handle
			_ = writeErr(rw, errors.New("missing channel in request"), http.StatusBadRequest)
			return
		}

		if err := s.botService.LeaveChannel(req.Channel); err != nil {
			// TODO: Handle
			_ = writeErr(rw, fmt.Errorf("failed to leave channel: %w", err), http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (s *server) BotInfo() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		botInfos := s.botService.BotInfo()
		if err := writeJSON(rw, botInfos, http.StatusOK); err != nil {
			s.logger.Error("failed to write JSON", zap.Error(err))
		}
	}
}

func (s *server) ChannelInfo() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		chanInfo := s.botService.ChannelInfo()
		if err := writeJSON(rw, chanInfo, http.StatusOK); err != nil {
			s.logger.Error("failed to write JSON", zap.Error(err))
		}
	}
}

// writeJSON writes a JSON payload back to the ResponseWriter with a status code
func writeJSON(rw http.ResponseWriter, payload interface{}, status int) error {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	if err := json.NewEncoder(rw).Encode(payload); err != nil {
		return fmt.Errorf("Encode: %w", err)
	}
	return nil
}

func writeErr(rw http.ResponseWriter, err error, status int) error {
	type errorResponse struct {
		Error string `json:"error"`
	}

	return writeJSON(rw, errorResponse{err.Error()}, status)
}
