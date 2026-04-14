.SILENT: export-server-env export-agent-env export-env

env-server:
	cp .env.agent.example .env.agent

env-agent:
	cp .env.agent.example .env.agent

env: env-agent env-server

build-server:
	go build -o ./cmd/server/server ./cmd/server/

build-agent:
	go build -o ./cmd/agent/agent ./cmd/agent/

build: build-server build-agent

run-server:
	@export $(shell cat .env.server | xargs) && go run ./cmd/server/main.go


run-agent:
	@export $(shell cat .env.agent | xargs) && go run ./cmd/agent/main.go
