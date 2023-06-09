package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/guiferpa/jackchain/blockchain"
)

var mu sync.Mutex

type Node struct {
	ID    string
	Port  string
	peers map[string]string
	chain *blockchain.Chain
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

func (n *Node) DisconnectFromPeers() error {
	message := []byte(fmt.Sprintf("%s %s", DISCONNECT, n.ID))
	if err := n.Broadcast(message); err != nil {
		return err
	}

	return nil
}

func (n *Node) AddPeer(id, host, port string) error {
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
	if _, ok := n.peers[id]; !ok {
		return errors.New("peer already removed")
	}

	delete(n.peers, id)

	return nil
}

type NodeHandler func(message, broadcast string) error

func (n *Node) SetHandler(t string, h NodeHandler) {}

func (n *Node) SetGenericHandler(h NodeHandler) {}

func (n *Node) Listen(port string) (<-chan []byte, <-chan []byte, error) {
	n.Port = port

	addr := fmt.Sprintf("0.0.0.0:%s", n.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	msgc := make(chan []byte)
	brdc := make(chan []byte)

	go read(ln, msgc)
	go write(n.ID, brdc)

	return msgc, brdc, nil
}

func NewNode(chain *blockchain.Chain) *Node {
	return &Node{ID: uuid.NewString(), chain: chain, peers: make(map[string]string, 0)}
}
