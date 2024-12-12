package peer

import (
	"context"
	"fmt"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/logger"
	protogreeter "github.com/guiferpa/jackiechain/proto/greeter"
)

type ID string

type Peer struct {
	ID                ID
	PeerHandshakeList map[ID]Peer
	Blockchain        *blockchain.Blockchain
	protogreeter.UnimplementedGreeterServer
}

func (s *Peer) ReachOut(ctx context.Context, pr *protogreeter.PingRequest) (*protogreeter.PongResponse, error) {
	logger.Yellow(fmt.Sprintf("Ping from agent %s", pr.Aid))
	return &protogreeter.PongResponse{Pid: string(s.ID)}, nil
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

func New(id ID, bc *blockchain.Blockchain) *Peer {
	return &Peer{ID: id, Blockchain: bc, PeerHandshakeList: make(map[ID]Peer, 0)}
}
