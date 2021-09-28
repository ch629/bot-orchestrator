package main

import (
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/bots"
	"github.com/ch629/irc-bot-orchestrator/internal/pkg/server"
)

// grpcurl -plaintext -import-path . -proto orchestrator.proto -d '{}' localhost:8080 Orchestrator/Join
func main() {
	// TODO: Setup logging
	botsService := bots.New()
	if err := server.New(*botsService).Start(8080); err != nil {
		panic(err)
	}
}
