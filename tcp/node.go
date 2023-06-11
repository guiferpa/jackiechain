package tcp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/guiferpa/jackiechain/blockchain"
)

var mu sync.Mutex

type NodeHandler func(net.Conn, bool) error

type NodeTerminateHandler func(os.Signal) (int, error)

type Node struct {
	ID               string
	UpAt             time.Time
	Chain            *blockchain.Chain
	Config           NodeConfig
	peers            map[string]string
	unconfirmedTxs   map[string]blockchain.Transaction
	handler          NodeHandler
	httpRouter       chi.Router
	terminateHandler NodeTerminateHandler
}

type NodeStats struct {
	ID       string        `json:"id"`
	Uptime   time.Duration `json:"uptime"`
	Peers    []interface{} `json:"peers"`
	NodePort string        `json:"node_port"`
}

func read(ln net.Listener, h NodeHandler, verbose bool) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		if err := h(conn, verbose); err != nil && err != io.EOF {
			red := color.New(color.FgBlack, color.BgHiRed).SprintFunc()
			log.Println(red(err))
		}
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
	mu.Lock()
	defer mu.Unlock()

	message := []byte(fmt.Sprintf("JACKIE %s %s 0.0.0.0 %s", JACKIE_CONNECT, n.ID, n.Config.NodePort))
	if err := send(fmt.Sprintf("0.0.0.0:%s", port), message); err != nil {
		return err
	}

	for key, peer := range n.peers {
		pport := strings.Split(peer, ":")[1]
		message = []byte(fmt.Sprintf("JACKIE %s %s 0.0.0.0 %s", JACKIE_CONNECT, key, pport))
		if err := send(fmt.Sprintf("0.0.0.0:%s", port), message); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) Connect(host, port string) error {
	message := []byte(fmt.Sprintf("JACKIE %s %s 0.0.0.0 %s", JACKIE_CONNECT, n.ID, n.Config.NodePort))
	return send(fmt.Sprintf("%s:%s", host, port), message)
}

func (n *Node) DisconnectPeers() error {
	message := []byte(fmt.Sprintf("JACKIE %s %s", JACKIE_DISCONNECT, n.ID))
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

func (n *Node) RequestTxApprobation(tx blockchain.Transaction) error {
	mu.Lock()
	defer mu.Unlock()

	if n.unconfirmedTxs == nil {
		n.unconfirmedTxs = make(map[string]blockchain.Transaction)
	}

	n.unconfirmedTxs[tx.CalculateHash()] = tx

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("JACKIE %s %s %s", JACKIE_TX_APPROBATION, n.ID, base64.StdEncoding.EncodeToString(txmsg))

	return n.Broadcast([]byte(msg))
}

func (n *Node) RequestTxApprobationOK(tx blockchain.Transaction, dstNid string) error {
	mu.Lock()
	defer mu.Unlock()

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("JACKIE %s %s %s %s", JACKIE_TX_APPROBATION_OK, n.ID, dstNid, base64.StdEncoding.EncodeToString(txmsg))

	return n.Broadcast([]byte(msg))
}

func (n *Node) RequestTxApprobationFail(tx blockchain.Transaction, dstNid string) error {
	mu.Lock()
	defer mu.Unlock()

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("JACKIE %s %s %s %s", JACKIE_TX_APPROBATION_FAIL, n.ID, dstNid, base64.StdEncoding.EncodeToString(txmsg))

	return n.Broadcast([]byte(msg))
}

func (n *Node) SetHandler(h NodeHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.handler = h
}

func (n *Node) SetTerminateHandler(h NodeTerminateHandler) {
	mu.Lock()
	defer mu.Unlock()

	n.terminateHandler = h
}

func (n *Node) Listen(sigc chan os.Signal) error {
	addr := fmt.Sprintf("0.0.0.0:%s", n.Config.NodePort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	connc := make(chan net.Conn)
	brdc := make(chan []byte)

	mu.Lock()
	handler := n.handler
	verbose := n.Config.Verbose
	mu.Unlock()

	go read(ln, handler, verbose)

	go write(n.ID, brdc)

	go func(connc chan net.Conn, brdc chan []byte) {
		for {
			select {
			case brd := <-brdc:
				if err := n.Broadcast([]byte(fmt.Sprintf("JACKIE %s %s", JACKIE_MESSAGE, string(brd)))); err != nil {
					panic(err)
				}

			case sig := <-sigc:
				mu.Lock()
				handler := n.terminateHandler
				mu.Unlock()

				code, err := handler(sig)
				if err != nil {
					fmt.Println(err)
				}

				os.Exit(code)
			}
		}
	}(connc, brdc)

	return nil
}

func (n *Node) Stats() NodeStats {
	peers := make([]interface{}, 0)

	mu.Lock()

	uptime := time.Now().Sub(n.UpAt)

	for id, peer := range n.peers {
		peers = append(peers, map[string]string{
			"id":      id,
			"address": strings.Trim(peer, "\x00"),
		})
	}

	mu.Unlock()

	return NodeStats{
		ID:       n.ID,
		Uptime:   time.Duration(uptime / time.Second),
		Peers:    peers,
		NodePort: n.Config.NodePort,
	}
}

type NodeConfig struct {
	NodePort string
	Verbose  bool
}

func NewNode(config NodeConfig, chain *blockchain.Chain) *Node {
	httpRouter := chi.NewRouter()

	return &Node{
		ID:         uuid.NewString(),
		UpAt:       time.Now(),
		Chain:      chain,
		httpRouter: httpRouter,
		peers:      make(map[string]string, 0),
		Config:     config,
	}
}
