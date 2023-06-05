build:
	CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jack ./cmd/cli/*.go
