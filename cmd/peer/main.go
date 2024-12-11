package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/guiferpa/jackiechain/block"
	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"
	"github.com/guiferpa/jackiechain/transaction"
	"google.golang.org/grpc"
)

func main() {
	serverPort := flag.Int("server-port", 9000, "server port")

	flag.Parse()

	bc := &blockchain.Blockchain{
		Blocks:           make(block.BlockMap),
		Txs:              make(transaction.TxMap),
		PendingTxs:       make(transaction.TxMap),
		MiningDifficulty: 4,
		UTxOs:            make(transaction.UTxOMap),
		GenesisBlock:     nil,
		LatestBlock:      nil,
	}

	peerID := uuid.New().String()

	logger.Magenta(fmt.Sprintf("Initializing peer %s", peerID))

	p := NewPeer(peerID, bc)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *serverPort))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	s := grpc.NewServer()

	proto.RegisterGreeterServer(s, p)

	logger.Magenta(fmt.Sprintf("Running gRPC server on port %v", *serverPort))

	go p.SetBuildBlockInterval(time.NewTicker(time.Second * 5))

	if err := s.Serve(listener); err != nil {
		logger.Red(err.Error())
		os.Exit(1)
	}
}
