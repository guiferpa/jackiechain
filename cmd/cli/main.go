package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	bc "github.com/guiferpa/jackchain/blockchain"
	inet "github.com/guiferpa/jackchain/net"
	"github.com/guiferpa/jackchain/wallet"
)

func read(ln net.Listener, msgc chan string, errc chan error) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			errc <- err
			continue
		}

		bs := make([]byte, 1024)
		if _, err := conn.Read(bs); err != nil {
			errc <- err
			continue
		}

		msgc <- string(bs)
	}
}

func write(cmdc chan string, errc chan error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			errc <- err
		}

		cmdc <- s
	}
}

type commanderOptions struct {
	Action   string
	Resource string
	Args     []string
}

func commander(opts commanderOptions, node *inet.Node, msgc chan string, errc chan error) {
	if strings.ToLower(opts.Action) == "create" {
		switch strings.ToLower(opts.Resource) {
		case "wallet":
			w, err := wallet.NewWallet()
			if err != nil {
				errc <- err
				break
			}

			if err = w.ExportPrivateKey(); err != nil {
				errc <- err
				break
			}

			msgc <- fmt.Sprintln("Wallet address:", w.GetAddress())

		case "tx":
			w, err := wallet.ParseWallet()
			if err != nil {
				errc <- err
				break
			}

			amount, err := strconv.ParseInt(opts.Args[1], 10, 64)
			if err != nil {
				errc <- err
				break
			}

			tx := bc.NewSignedTransaction(bc.TransactionOptions{
				Sender:       *w,
				ReceiverAddr: opts.Args[0],
				Amount:       amount,
			})
			if err := node.Chain.AddTransaction(tx); err != nil {
				errc <- err
				break
			}

			msgc <- fmt.Sprintln("Transaction", tx.CalculateHash(), "created")

		default:
			errc <- errors.New(fmt.Sprintf("Resource %s not found", opts.Resource))
		}

	}
}

func main() {
	network := "tcp"
	port := "3000"
	address := fmt.Sprintf("0.0.0.0:%s", port)

	ln, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	chain := bc.NewChain(bc.ChainOptions{})
	node := inet.NewNode(address, chain)

	log.Println("Node's running at", fmt.Sprintf("%s/%s", network, address))

	cmdc := make(chan string, 0)
	msgc := make(chan string, 0)
	errc := make(chan error, 0)

	go read(ln, msgc, errc)
	go write(cmdc, errc)

	for {
		select {
		case err := <-errc:
			log.Println(err)

		case msg := <-msgc:
			fmt.Print("Message:", string(msg))

		case cmd := <-cmdc:
			cmd = strings.Trim(cmd, string('\n'))

			if cmd == "" {
				continue
			}

			l := strings.Fields(cmd)

			if len(l) < 2 {
				fmt.Println("Command", cmd, "is wrong")
				continue
			}

			opts := commanderOptions{
				Action:   l[0],
				Resource: l[1],
				Args:     l[2:],
			}

			go commander(opts, node, msgc, errc)
		}
	}
}
