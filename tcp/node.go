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
	unconfirmedTxs   map[string][]string
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

func Send(addr string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}

func SendJackieRequest(addr string, act, msg string) error {
	return Send(addr, []byte(fmt.Sprintf("JACKIE %s %s", act, msg)))
}

func (n *Node) Broadcast(act, msg string) error {
	for _, peer := range n.peers {
		if err := SendJackieRequest(peer, act, msg); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) ShareConnectionState(host, port string) error {
	mu.Lock()
	defer mu.Unlock()

	addr := fmt.Sprintf("%s:%s", host, port)
	message := fmt.Sprintf("%s %s %s", n.ID, "0.0.0.0", n.Config.NodePort)
	if err := SendJackieRequest(addr, JACKIE_CONNECT_LOOPBACK, message); err != nil {
		return err
	}

	for key, peer := range n.peers {
		pport := strings.Split(peer, ":")[1]
		addr = fmt.Sprintf("0.0.0.0:%s", port)
		message = fmt.Sprintf("%s %s %s", key, "0.0.0.0", pport)
		if err := SendJackieRequest(addr, JACKIE_CONNECT, message); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) Connect(host, port string) error {
	message := []byte(fmt.Sprintf("JACKIE %s %s 0.0.0.0 %s", JACKIE_CONNECT, n.ID, n.Config.NodePort))
	return Send(fmt.Sprintf("%s:%s", host, port), message)
}

func (n *Node) ConnectOK(key string) error {
	mu.Lock()
	defer mu.Unlock()

	if peer, ok := n.peers[key]; ok {
		message := fmt.Sprintf("%s", n.ID)
		return SendJackieRequest(peer, JACKIE_CONNECT_OK, message)
	}

	return errors.New(fmt.Sprintf("unheathly node, it wasn't possible download blockchaim from %s", key))
}

func (n *Node) DisconnectPeers() error {
	message := fmt.Sprintf("%s", n.ID)
	if err := n.Broadcast(JACKIE_DISCONNECT, message); err != nil {
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
		n.unconfirmedTxs = make(map[string][]string)
	}

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	for id, peer := range n.peers {
		n.unconfirmedTxs[tx.CalculateHash()] = append(n.unconfirmedTxs[tx.CalculateHash()], id)

		message := fmt.Sprintf("%s %s", n.ID, base64.StdEncoding.EncodeToString(txmsg))
		if err := SendJackieRequest(peer, JACKIE_TX_APPROBATION, message); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) RequestTxApprobationOK(tx blockchain.Transaction, peerid string) error {
	mu.Lock()
	defer mu.Unlock()

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	if addr, ok := n.peers[peerid]; ok {
		message := fmt.Sprintf("%s %s %s", n.ID, peerid, base64.StdEncoding.EncodeToString(txmsg))
		if err := SendJackieRequest(addr, JACKIE_TX_APPROBATION_OK, message); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) RequestTxApprobationFail(tx blockchain.Transaction, dstNid string) error {
	mu.Lock()
	defer mu.Unlock()

	txmsg, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("%s %s %s", n.ID, dstNid, base64.StdEncoding.EncodeToString(txmsg))
	return n.Broadcast(JACKIE_TX_APPROBATION_FAIL, message)
}

func (n *Node) UploadBlockchainTo(key string) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	buf, err := json.Marshal(n.Chain)
	if err != nil {
		return 0, err
	}

	key = string(bytes.Trim([]byte(key), "\x00"))
	if peer, ok := n.peers[key]; ok {
		message := fmt.Sprintf("%s %s %s", n.ID, key, base64.StdEncoding.EncodeToString(buf))
		return len(buf), SendJackieRequest(peer, JACKIE_DOWNLOAD_BLOACKCHAIN_OK, message)
	}

	return 0, errors.New(fmt.Sprintf("there's no peer with key equals %s to transfer blockchain's state", key))
}

func (n *Node) CommitTxApproved(jury string, tx *blockchain.Transaction) error {
	peers := n.unconfirmedTxs[tx.CalculateHash()]

	npeers := make([]string, 0)
	for _, peer := range peers {
		if peer != jury {
			npeers = append(npeers, peer)
		}
	}

	if len(npeers) == 0 {
		if err := n.Chain.AddTransaction(tx); err != nil {
			return err
		}

		delete(n.unconfirmedTxs, tx.CalculateHash())
		return nil
	}

	n.unconfirmedTxs[tx.CalculateHash()] = npeers

	return nil
}

func (n *Node) RequestDownloadBlockchain(host, port string) error {
	mu.Lock()
	defer mu.Unlock()

	message := fmt.Sprintf("%s", n.ID)
	addr := fmt.Sprintf("%s:%s", host, port)
	if err := SendJackieRequest(addr, JACKIE_DOWNLOAD_BLOACKCHAIN, message); err != nil {
		return err
	}

	return nil
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
				message := fmt.Sprintf("%s", string(brd))
				if err := n.Broadcast(JACKIE_MESSAGE, message); err != nil {
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
	mu.Lock()
	defer mu.Unlock()

	peers := make([]interface{}, 0)

	uptime := time.Now().Sub(n.UpAt)

	for id, peer := range n.peers {
		peers = append(peers, map[string]string{
			"id":      id,
			"address": strings.Trim(peer, "\x00"),
		})
	}

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
