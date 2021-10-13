proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/orchestrator.proto

build:
	go build -o bin/app cmd/server/main.go


run: build
	./bin/app

build-client:
	go build -o bin/client cmd/client/main.go

run-client: build-client
	./bin/client

generate:
	go generate ./...

test:
	go test ./...
