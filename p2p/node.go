package p2p

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/google/uuid"

	"github.com/guiferpa/jackiechain/blockchain"
)

var mu sync.Mutex

type NodeHandler func(message []string) error

type NodeTerminateHandler func(os.Signal) (int, error)

type Node struct {
	ID               string
	Port             string
	peers            map[string]string
	chain            *blockchain.Chain
	handlers         map[string]NodeHandler
	terminateHandler NodeTerminateHandler
}

func read(ln net.Listener, msgc chan []byte) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		bs := make([]byte, 1024)
		if _, err := conn.Read(bs); err != nil {
			return err
		}

		msgc <- bs
	}
}

func write(nodeID string, brdc chan []byte) {
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		brdc <- []byte(fmt.Sprintf("%s: %s", nodeID, strings.Trim(s, string('\n'))))
	}
}

func send(addr string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}

func (n *Node) Broadcast(msg []byte) error {
	for _, peer := range n.peers {
		if err := send(peer, msg); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) ShareConnectionState(host, port string) error {
	message := []byte(fmt.Sprintf("%s %s %s %s", CONNECT, n.ID, "0.0.0.0", n.Port))
	if err := send(fmt.Sprintf("0.0.0.0:%s", port), message); err != nil {
		return err
	}

	for key, peer := range n.peers {
		pport := strings.Split(peer, ":")[1]
		message = []byte(fmt.Sprintf("%s %s %s %s", CONNECT, key, "0.0.0.0", pport))
		if err := send(fmt.Sprintf("0.0.0.0:%s", port), message); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) Connect(host, port string) error {
	message := []byte(fmt.Sprintf("%s %s %s %s", CONNECT, n.ID, "0.0.0.0", n.Port))
	return send(fmt.Sprintf("%s:%s", host, port), message)
}

func (n *Node) DisconnectPeers() error {
	message := []byte(fmt.Sprintf("%s %s", DISCONNECT, n.ID))
	if err := n.Broadcast(message); err != nil {
		return err
	}

	return nil
}

func (n *Node) AddPeer(id, host, port string) error {
	mu.Lock()
	defer mu.Unlock()

	if id == n.ID {
		return errors.New("it's not possible add itself")
	}

	if _, ok := n.peers[id]; ok {
		return errors.New("peer already added")
	}

	n.peers[id] = fmt.Sprintf("%s:%s", host, port)

	return nil
}

func (n *Node) RemovePeer(id string) error {
	mu.Lock()
	defer mu.Unlock()

	id = string(bytes.Trim([]byte(id), "\x00"))

	if _, ok := n.peers[id]; !ok {
		return errors.New("peer already removed")
	}

	delete(n.peers, id)

	return nil
}

func (n *Node) SetHandler(t string, h NodeHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.handlers[t] = h
}

func (n *Node) SetGenericHandler(h NodeHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.handlers["generic"] = h
}

func (n *Node) SetWriteHandler(h NodeHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.handlers["write"] = h
}

func (n *Node) SetTerminateHandler(h NodeTerminateHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.terminateHandler = h
}

func (n *Node) Listen(port string, verbose bool, sigc chan os.Signal) error {
	n.Port = port

	addr := fmt.Sprintf("0.0.0.0:%s", n.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	msgc := make(chan []byte)
	brdc := make(chan []byte)

	go read(ln, msgc)
	go write(n.ID, brdc)

	go func(msgc, brdc chan []byte) {
		for {
			select {
			case msg := <-msgc:
				if verbose {
					yellow := color.New(color.FgYellow).SprintFunc()
					log.Println(yellow(string(msg)))
				}

				line := strings.Fields(string(msg))

				mu.Lock()
				handler, exists := n.handlers[line[0]]
				if !exists {
					handler = n.handlers["generic"]
				}
				mu.Unlock()

				if err := handler(line); err != nil {
					panic(err)
				}

			case brd := <-brdc:
				if err := n.Broadcast(brd); err != nil {
					panic(err)
				}

			case sig := <-sigc:
				mu.Lock()
				code, err := n.terminateHandler(sig)
				mu.Unlock()

				if err != nil {
					fmt.Println(err)
				}

				os.Exit(code)
			}
		}
	}(msgc, brdc)

	return nil
}

func NewNode(chain *blockchain.Chain) *Node {
	return &Node{
		ID:       uuid.NewString(),
		chain:    chain,
		peers:    make(map[string]string, 0),
		handlers: make(map[string]NodeHandler, 0),
	}
}
