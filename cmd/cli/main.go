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
	"github.com/guiferpa/jackiechain/v2/node"
)

var (
	connect    string
	port       string
	verbose    bool
	walletAddr string
)

func init() {
	flag.StringVar(&connect, "connect", "", "set connection")
	flag.StringVar(&port, "port", "3000", "set port")
	flag.BoolVar(&verbose, "verbose", false, "set verbose")
	flag.StringVar(&walletAddr, "wallet", "", "set wallet address")
}

func main() {
	flag.Parse()

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

	miningTicker := time.NewTicker(1 * time.Minute)
	go node.MineNewBlock(walletAddr, chain, miningTicker)

	<-sigc

	node.TerminatePeer(peer)

	log.Println("Peer", peer.GetID(), "terminated")
}
