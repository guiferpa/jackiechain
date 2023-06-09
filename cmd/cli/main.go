package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/guiferpa/jackchain/blockchain"
	"github.com/guiferpa/jackchain/p2p"
)

var (
	peer    string
	port    string
	verbose bool
)

func init() {
	flag.StringVar(&peer, "peer", "", "set peer")
	flag.StringVar(&port, "port", "3000", "set port")
	flag.BoolVar(&verbose, "verbose", false, "set verbose")
}

var mu sync.Mutex

func main() {
	flag.Parse()

	chain := blockchain.NewChain(blockchain.ChainOptions{})
	node := p2p.NewNode(chain)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	if err := node.Listen(port, verbose, sigc); err != nil {
		panic(err)
	}

	if peer != "" {
		if err := node.Connect("0.0.0.0", peer); err != nil {
			panic(err)
		}
	}

	log.Println("Node", node.ID, "is running at", node.Port)

	node.SetHandler(p2p.CONNECT, func(message []string) error {
		if err := node.AddPeer(message[1], message[2], message[3]); err != nil {
			if err.Error() == "peer already added" {
				return nil
			}

			if err.Error() == "it's not possible add itself" {
				return nil
			}

			return err
		}

		log.Println("Connected to", message[1])

		if err := node.ShareConnectionState(message[2], message[3]); err != nil {
			return err
		}

		return nil
	})

	node.SetHandler(p2p.DISCONNECT, func(message []string) error {
		if err := node.RemovePeer(message[1]); err != nil {
			return err
		}

		log.Println("Disconnected to", message[1])

		return nil
	})

	node.SetGenericHandler(func(message []string) error {
		log.Printf(strings.Join(message, " "))

		return nil
	})

	node.SetTerminateHandler(func(sig os.Signal) (int, error) {
		log.Println("Node", node.ID, "is terminated")

		if err := node.DisconnectPeers(); err != nil {
			return 1, err
		}

		return 0, nil
	})

	select {}
}
