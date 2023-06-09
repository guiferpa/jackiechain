package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/httputil"
	"github.com/guiferpa/jackiechain/tcp"
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

	nodecfg := tcp.NodeConfig{
		NodePort: port,
		Verbose:  verbose,
	}
	chain := blockchain.NewChain(blockchain.ChainOptions{})
	node := tcp.NewNode(nodecfg, chain)

	node.SetHandler(func(conn net.Conn) error {
		defer conn.Close()

		if act, args, err := tcp.ParseJackieRequest(conn); err == nil {
			switch act {
			case tcp.JACKIE_CONNECT:
				if err := node.AddPeer(args[0], args[1], args[2]); err != nil {
					if err.Error() == "peer already added" {
						return nil
					}

					if err.Error() == "it's not possible add itself" {
						return nil
					}

					return err
				}

				log.Println("Connected to", args[0])

				if err := node.ShareConnectionState(args[1], args[2]); err != nil {
					return err
				}

				return nil

			case tcp.JACKIE_DISCONNECT:
				if err := node.RemovePeer(args[0]); err != nil {
					return err
				}

				log.Println("Disconnected to", args[0])

				return nil

			case tcp.JACKIE_MESSAGE:
				log.Print(strings.Join(args, " "))
			}
		} else {
			req, err := tcp.ParseHTTPRequest(conn)
			if err != nil {
				return err
			}

			resp := httputil.NewHTTPResponse(req, http.StatusOK, []byte(req.Method))

			_, err = httputil.Response(conn, resp)

			return err
		}

		return nil
	})

	node.SetTerminateHandler(func(sig os.Signal) (int, error) {
		log.Println("Node", node.ID, "is terminated")

		if err := node.DisconnectPeers(); err != nil {
			return 1, err
		}

		return 0, nil
	})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	if err := node.Listen(sigc); err != nil {
		panic(err)
	}

	log.Println("Node ID:", node.ID)
	log.Println("Node is running at", node.Config.NodePort)

	if peer != "" {
		if err := node.Connect("0.0.0.0", peer); err != nil {
			panic(err)
		}
	}

	select {}
}
