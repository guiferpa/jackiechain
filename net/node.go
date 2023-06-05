package net

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/guiferpa/jackchain/blockchain"
)

var PROTOCOL = "tcp"

type Node struct {
	Addr        string
	FingerTable map[string]Finger
	Chain       *blockchain.Chain
}

func (n *Node) Listen(doneCh chan bool, errCh chan error) {
	l, err := net.Listen(PROTOCOL, n.Addr)
	if err != nil {
		errCh <- err
		return
	}

	defer l.Close()

	fmt.Println("[jackie] Node listening on", n.Addr, "using protocol", PROTOCOL)

	doneCh <- true

	for {
		conn, err := l.Accept()
		if err != nil {
			errCh <- err
			return
		}

		buf := make([]byte, 1024)
		if _, err := conn.Read(buf); err != nil {
			errCh <- err
			return
		}

		message := strings.Split(string(buf), " ")

		if status := message[0]; status == "OK" {
			switch op := message[1]; op {
			case "TX":
				fmt.Println(message)

				amount, err := strconv.ParseInt(message[4], 10, 64)
				if err != nil {
					errCh <- err
				}

				tx := &blockchain.Transaction{
					Sender:    message[2],
					Receiver:  message[3],
					Amount:    amount,
					Signature: []byte(message[5]),
				}

				if err := n.Chain.AddTransaction(tx); err != nil {
					errCh <- err
				}

			case "CONNECT":
				id := message[3]
				if _, ok := n.FingerTable[id]; !ok {
					fmt.Println(message)

					finger := NewFinger(message[2])
					n.FingerTable = map[string]Finger{message[3]: finger}

					dial, err := net.Dial(PROTOCOL, message[2])
					if err != nil {
						errCh <- err
						return
					}

					h := md5.New()
					io.WriteString(h, n.Addr)

					if _, err := dial.Write([]byte(fmt.Sprintf("OK CONNECT %s %x", n.Addr, h.Sum(nil)))); err != nil {
						errCh <- err
						return
					}
				}

			case "DISCONNECT":
				fmt.Println(message)
				id := message[3]
				delete(n.FingerTable, id)
			}
		}

		conn.Close()
	}
}

func (n *Node) Join(addr string, doneCh chan bool, errCh chan error) {
	for {
		select {
		case done := <-doneCh:
			if done {
				conn, err := net.Dial(PROTOCOL, addr)
				if err != nil {
					errCh <- err
					return
				}

				h := md5.New()
				io.WriteString(h, n.Addr)

				if _, err := conn.Write([]byte(fmt.Sprintf("OK CONNECT %s %x", n.Addr, h.Sum(nil)))); err != nil {
					errCh <- err
					return
				}

				fmt.Println("[jackie] Myself joined at", addr, "using protocol", PROTOCOL)
			}
		}
	}
}

func (n *Node) Disconnect(addr string) error {
	conn, err := net.Dial(PROTOCOL, addr)
	if err != nil {
		return err
	}

	h := md5.New()
	io.WriteString(h, n.Addr)

	if _, err := conn.Write([]byte(fmt.Sprintf("OK DISCONNECT %s %x", n.Addr, h.Sum(nil)))); err != nil {
		return err
	}

	return nil
}

func NewNode(addr string, chain *blockchain.Chain) (*Node, error) {
	return &Node{
		Addr:        addr,
		Chain:       chain,
		FingerTable: make(map[string]Finger, 0),
	}, nil
}
