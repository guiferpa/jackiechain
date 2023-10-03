package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"
)

type Server struct {
	proto.UnimplementedGreeterServer
	Blockchain *blockchain.Blockchain
}

func (s *Server) ReachOut(ctx context.Context, pr *proto.PingRequest) (*proto.PongResponse, error) {
	return &proto.PongResponse{Text: "PONG"}, nil
}

func (s *Server) SetBuildBlockInterval(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			bh, err := blockchain.BuildBlock(s.Blockchain)
			if err != nil {
				logger.Red(err.Error())
				continue
			}
			logger.Magenta(fmt.Sprintf("Block %s was built", bh))

		}
	}
}

func NewServer(bc *blockchain.Blockchain) *Server {
	return &Server{Blockchain: bc}
}
