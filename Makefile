build:
	CGO_ENABLED=0 go build -o $(shell go env GOPATH)/bin/jack ./cmd/cli/*.go
