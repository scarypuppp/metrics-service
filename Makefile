build-server:
	go build -o ./cmd/server/server ./cmd/server/
build-agent:
	go build -o ./cmd/agent/agent ./cmd/agent/

build:
	go build -o ./cmd/agent/agent ./cmd/agent/
	go build -o ./cmd/server/server ./cmd/server/

run-server:
	go run ./cmd/server/main.go

run-agent:
	go run ./cmd/agent/main.go