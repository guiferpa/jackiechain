package main

import (
	"context"

	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	client := proto.NewGreeterClient(conn)
	resp, err := client.ReachOut(context.Background(), &proto.PingRequest{})
	if err != nil {
		logger.Red(err.Error())
		return
	}
	logger.Yellow(resp.Text)
}
