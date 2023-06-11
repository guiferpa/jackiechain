package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/httputil"
	"github.com/guiferpa/jackiechain/tcp"
	"github.com/guiferpa/jackiechain/wallet"
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

	node.SetHandler(func(conn net.Conn, verbose bool) error {
		defer conn.Close()

		message := make([]byte, 1024)
		if _, err := conn.Read(message); err != nil {
			return err
		}

		if verbose {
			yellow := color.New(color.FgYellow).SprintFunc()
			log.Println(yellow(string(message)))
		}

		if act, args, err := tcp.ParseJackieRequest(bytes.NewBuffer(message)); err == nil {
			switch act {
			case tcp.JACKIE_TX_APPROBATION_OK:
				jury := args[0]
				// origin := args[1]
				etx := strings.Trim(args[2], "\x00")

				btx, err := base64.StdEncoding.DecodeString(etx)
				if err != nil {
					return err
				}

				tx := blockchain.Transaction{}
				if err := json.Unmarshal(btx, &tx); err != nil {
					return err
				}

				if err := node.CommitTxApproved(jury, &tx); err != nil {
					return err
				}

				log.Println("Transaction", tx.CalculateHash(), "approved by node", jury)

			case tcp.JACKIE_TX_APPROBATION:
				nid := args[0]
				etx := strings.Trim(args[1], "\x00")

				btx, err := base64.StdEncoding.DecodeString(etx)
				if err != nil {
					return err
				}

				tx := blockchain.Transaction{}
				if err := json.Unmarshal(btx, &tx); err != nil {
					return err
				}

				if err := node.Chain.AddTransaction(&tx); err != nil {
					return node.RequestTxApprobationFail(tx, nid)
				}

				if err := node.RequestTxApprobationOK(tx, nid); err != nil {
					return err
				}

				log.Println("Transaction", tx.CalculateHash(), "approved")

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
			req, err := tcp.ParseHTTPRequest(bytes.NewBuffer(message))
			if err != nil {
				return err
			}

			if req.Method == http.MethodGet {
				switch req.URL.Path {
				case "/chain":
					buf := &bytes.Buffer{}
					if err := json.NewEncoder(buf).Encode(node.Chain); err != nil {
						return err
					}
					return httputil.Response(req, conn, http.StatusOK, buf)

				case "/stats":
					buf := &bytes.Buffer{}
					if err := json.NewEncoder(buf).Encode(node.Stats()); err != nil {
						return err
					}
					return httputil.Response(req, conn, http.StatusOK, buf)

					/***
					TODO: Created router to route these resources
					In this GET - /wallets I'd like to receive
					private seed as url params instead of query params

					Start: https://dev.to/bmf_san/introduction-to-golang-http-router-made-with-nethttp-3nmb
					**/

				case "/wallets":
					seed := req.URL.Query().Get("seed")

					if seed == "" {
						return httputil.ResponseNotFound(req, conn)
					}

					w, err := wallet.ParseWallet(seed)
					if err != nil {
						return err
					}

					payload := map[string]interface{}{
						"address":      w.GetAddress(),
						"private_seed": w.GetPrivateSeed(),
					}

					buf := &bytes.Buffer{}
					if err := json.NewEncoder(buf).Encode(payload); err != nil {
						return err
					}
					return httputil.Response(req, conn, http.StatusOK, buf)

				default:
					return httputil.ResponseNotFound(req, conn)
				}
			}

			if req.Method == http.MethodPost {
				switch req.URL.Path {
				case "/tx":
					body := struct {
						Sender      string `json:"sender"`
						Receiver    string `json:"receiver"`
						Amount      int64  `json:"amount"`
						PrivateSeed string `json:"private_seed"`
					}{}

					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						return err
					}
					defer req.Body.Close()

					w, err := wallet.ParseWallet(body.PrivateSeed)
					if err != nil {
						return err
					}

					if body.Sender != w.GetAddress() {
						return httputil.ResponseBadRequest(req, conn, "Invalid wallet sender address")
					}

					tx := blockchain.NewSignedTransaction(blockchain.TransactionOptions{
						Sender:       *w,
						ReceiverAddr: body.Receiver,
						Amount:       body.Amount,
					})

					if err := node.RequestTxApprobation(*tx); err != nil {
						return err
					}

					return httputil.Response(req, conn, http.StatusNoContent, nil)

				case "/wallets":
					w, err := wallet.NewWallet()
					if err != nil {
						return err
					}

					payload := map[string]interface{}{
						"address":      w.GetAddress(),
						"private_seed": w.GetPrivateSeed(),
					}

					buf := &bytes.Buffer{}
					if err := json.NewEncoder(buf).Encode(payload); err != nil {
						return err
					}
					return httputil.Response(req, conn, http.StatusCreated, buf)

				default:
					return httputil.ResponseNotFound(req, conn)
				}
			}
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

	<-sigc
}
