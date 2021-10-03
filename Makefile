proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/orchestrator.proto

build:
	go build -o app cmd/main.go


run: build
	./app
