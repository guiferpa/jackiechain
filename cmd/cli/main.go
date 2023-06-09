package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"

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

	msgc, brdc, err := node.Listen(port)
	if err != nil {
		panic(err)
	}

	if peer != "" {
		if err := node.Connect("0.0.0.0", peer); err != nil {
			panic(err)
		}
	}

	log.Println("Node", node.ID, "is running at", node.Port)

	node.SetHandler(p2p.CONNECT, func(message, broadcast string) error {
		return nil
	})

	node.SetHandler(p2p.DISCONNECT, func(message, broadcast string) error {
		return nil
	})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	for {
		select {
		case msg := <-msgc:
			if verbose {
				yellow := color.New(color.FgYellow).SprintFunc()
				log.Println(yellow(string(msg)))
			}

			line := strings.Fields(string(msg))

			switch line[0] {
			case p2p.CONNECT:
				mu.Lock()

				err := node.AddPeer(line[1], line[2], line[3])

				mu.Unlock()

				if err != nil {
					if err.Error() == "peer already added" {
						continue
					}

					if err.Error() == "it's not possible add itself" {
						continue
					}

					panic(err)
				}

				log.Println("Connected to", line[1])

				if err := node.ShareConnectionState(line[2], line[3]); err != nil {
					panic(err)
				}

			case p2p.DISCONNECT:
				mu.Lock()

				err := node.RemovePeer(line[1])

				mu.Unlock()

				if err != nil {
					panic(err)
				}

				log.Println("Disconnected to", line[1])

			default:
				log.Printf(string(msg))
			}

		case brd := <-brdc:
			if err := node.Broadcast(brd); err != nil {
				panic(err)
			}

		case <-sigc:
			if err := node.DisconnectFromPeers(); err != nil {
				panic(err)
			}
			log.Println("Node", node.ID, "is terminated")
			os.Exit(0)
		}
	}
}
