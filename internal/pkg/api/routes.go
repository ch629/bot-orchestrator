package api

import "github.com/gorilla/mux"

func (s *server) createRoutes() *mux.Router {
	router := mux.NewRouter()

	subrouter := router.PathPrefix("/api/v1").Subrouter()
	subrouter.HandleFunc("/join", s.JoinChannel()).Methods("POST")
	subrouter.HandleFunc("/leave", s.LeaveChannel()).Methods("POST")
	subrouter.HandleFunc("/bot", s.BotInfo()).Methods("GET")
	subrouter.HandleFunc("/channel", s.ChannelInfo()).Methods("GET")
	return router
}
