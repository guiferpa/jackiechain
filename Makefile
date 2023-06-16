install: build

build:
	@CGO_ENABLED=1 go build -race -o $(shell go env GOPATH)/bin/jackie ./cmd/cli/*.go

prod:
	@CGO_ENABLED=0 go build -v -o jackie ./cmd/cli/*.go

