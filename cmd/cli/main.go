package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/v2/logger"
	"github.com/guiferpa/jackiechain/v2/node"
	"github.com/guiferpa/jackiechain/wallet"
)

var (
	connect string
	port    string
	verbose bool
	miner   string
)

func init() {
	flag.StringVar(&connect, "connect", "", "set connection")
	flag.StringVar(&port, "port", "3000", "set port")
	flag.BoolVar(&verbose, "verbose", false, "set verbose")
	flag.StringVar(&miner, "wallet", "", "set miner wallet address")
}

func main() {
	flag.Parse()

	if miner == "" {
		logger.Magenta("Miner wallet flag is empty, by default jackie will created new one")

		w, err := wallet.NewWallet()
		if err != nil {
			panic(err)
		}

		log.Println("Wallet address:", w.GetAddress())
		log.Println("Wallet private seed:", w.GetPrivateSeed())

		miner = w.GetAddress()
	}

	chain := blockchain.NewChain(blockchain.ChainOptions{
		MiningDifficulty: 2,
		MiningReward:     10,
	})
	peer := node.NewService(port)

	log.Println("Peer ID:", peer.GetID())

	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go node.Listen(port, verbose, peer, chain)

	log.Println("Node is running at", port)

	if connect != "" {
		conn, err := net.Dial("udp", connect)
		if err != nil {
			panic(err)
		}
		ip := conn.LocalAddr().(*net.UDPAddr).IP
		addr := fmt.Sprintf("%s:%s", ip, port)
		if err := node.PeerConnectRequest(peer.GetID(), addr, connect); err != nil {
			panic(err)
		}
	}

	miningTicker := time.NewTicker(2 * time.Minute)
	go node.MineNewBlock(miner, chain, miningTicker)

	<-sigc

	node.TerminatePeer(peer)

	log.Println("Peer", peer.GetID(), "terminated")
}
