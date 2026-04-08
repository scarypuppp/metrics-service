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

export-agent-env:
	export $(cat .env.agent | xargs)

export-server-env:
	export $(cat .env.server | xargs)

export-env: export-server-env export-agent-env

run-server:
	$(MAKE) export-server-env
	go run ./cmd/server/main.go

run-agent:
	$(MAKE) export-agent-env
	go run ./cmd/agent/main.go
