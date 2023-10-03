package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/guiferpa/jackiechain/block"
	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"
	"github.com/guiferpa/jackiechain/transaction"

	"google.golang.org/grpc"
)

func main() {
	bc := &blockchain.Blockchain{
		Blocks:           make(block.BlockMap),
		Txs:              make(transaction.TxMap),
		PendingTxs:       make(transaction.TxMap),
		MiningDifficulty: 4,
		UTxOs:            make(transaction.UTxOMap),
		GenesisBlock:     nil,
		LatestBlock:      nil,
	}

	is := NewServer(bc)

	logger.Magenta(fmt.Sprint("Running gRPC server"))

	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		logger.Red(err.Error())
		return
	}

	s := grpc.NewServer()

	proto.RegisterGreeterServer(s, is)

	go is.SetBuildBlockInterval(time.NewTicker(time.Second * 5))

	if err := s.Serve(listener); err != nil {
		logger.Red(err.Error())
		os.Exit(1)
	}
}
