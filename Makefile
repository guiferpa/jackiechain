install: build_server build_client

build_server:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-server ./cmd/server/*.go

build_client:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie-client ./cmd/client/*.go

