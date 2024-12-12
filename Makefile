install: peer agent

peer:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-peer ./cmd/peer/*.go

agent:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-agent ./cmd/agent/*.go

proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/**/*.proto

.PHONY: peer agent proto
