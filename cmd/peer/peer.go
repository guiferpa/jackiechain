package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"
)

type Peer struct {
	ID string
	proto.UnimplementedGreeterServer
	Blockchain *blockchain.Blockchain
}

func (s *Peer) ReachOut(ctx context.Context, pr *proto.PingRequest) (*proto.PongResponse, error) {
	logger.Yellow("PING")
	return &proto.PongResponse{Text: "PONG"}, nil
}

func (s *Peer) SetBuildBlockInterval(ticker *time.Ticker) {
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

func NewPeer(id string, bc *blockchain.Blockchain) *Peer {
	return &Peer{ID: id, Blockchain: bc}
}
