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
	"github.com/guiferpa/jackiechain/logger"
	"github.com/guiferpa/jackiechain/peer"
	"github.com/guiferpa/jackiechain/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverPort := flag.Int("server-port", 9000, "server port")
	nodeRemote := flag.String("node-remote", "", "node remote (no standalone config)")

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

	p := peer.New(peer.ID(peerID), bc)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *serverPort))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	cherr := make(chan error)
	serving := make(chan struct{})
	go p.Serve(listener, *nodeRemote, serving, cherr)
	<-serving
	logger.Magenta(fmt.Sprintf("Running gRPC server on port %v", *serverPort))

	go p.SetBuildBlockInterval(time.NewTicker(time.Second * 5))

	if *nodeRemote != "" {
		conn, err := grpc.NewClient(*nodeRemote, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Red(err.Error())
			os.Exit(2)
		}
		if err := p.TryConnect(conn); err != nil {
			logger.Red(err.Error())
			os.Exit(3)
		}
	}

	err = <-cherr
	logger.Red(err.Error())
	os.Exit(1)
}
