package net

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"strings"
)

var PROTOCOL = "tcp"

type Node struct {
	Addr        string
	FingerTable map[string]Finger
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

		message := strings.Fields(string(buf))

		if message[0] == "OK" {
			switch message[1] {
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

func NewNode(addr string) (*Node, error) {
	return &Node{
		Addr:        addr,
		FingerTable: make(map[string]Finger, 0),
	}, nil
}
