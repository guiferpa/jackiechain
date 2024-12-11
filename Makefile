install: build_peer build_agent

build_peer:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-peer ./cmd/peer/*.go

build_agent:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-agent ./cmd/agent/*.go

