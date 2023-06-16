package node

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/httputil"
)

var mu sync.Mutex

type Peer interface {
	GetID() string
	GetNeighborhood() map[string]string
	GetNeighborByID(id string) string
	SetNeighbor(id, addr string) error
	UnsetNeighbor(id string)
}

type Service struct {
	id           string
	neighborhood map[string]string
	verbose      bool
}

func (s *Service) GetID() string {
	return s.id
}

func (s *Service) SetNeighbor(id, addr string) error {
	id = strings.Trim(id, "\x00")
	addr = strings.Trim(addr, "\x00")

	if _, exists := s.neighborhood[id]; exists {
		ErrJackieDuplcatedPeer.PeerId = id
		ErrJackieDuplcatedPeer.PeerAddr = addr
		return ErrJackieDuplcatedPeer
	}

	s.neighborhood[id] = addr

	return nil
}

func (s *Service) GetNeighborByID(id string) string {
	return s.neighborhood[strings.Trim(id, "\x00")]
}

func (s *Service) UnsetNeighbor(id string) {
	delete(s.neighborhood, id)
}

func (s *Service) GetNeighborhood() map[string]string {
	return s.neighborhood
}

func JackieHandler(peer Peer, chain *blockchain.Chain, upat time.Time, port, action string, args []string) error {
	mu.Lock()
	defer mu.Unlock()

	// CONNECT <peer-id> <peer-addr>
	if action == JACKIE_CONNECT {
		if err := PeerConnectLoopback(peer, port, args[1]); err != nil {
			return err
		}

		for _, neighbor := range peer.GetNeighborhood() {
			if err := PeerConnectRequest(args[0], args[1], neighbor); err != nil {
				return err
			}
		}

		if err := peer.SetNeighbor(args[0], args[1]); err != nil {
			if errors.Is(err, ErrJackieDuplcatedPeer) {
				return nil
			}
			return err
		}

		log.Println("Connected to", args[0])

		return nil
	}

	// CONNECT_LOOPBACK <peer-id> <peer-addr>
	if action == JACKIE_CONNECT_LOOPBACK {
		if err := peer.SetNeighbor(args[0], args[1]); err != nil {
			if errors.Is(err, ErrJackieDuplcatedPeer) {
				return nil
			}
			return err
		}

		log.Println("Connected to", args[0])

		if err := DownloadBlockchainRequest(peer.GetID(), args[1]); err != nil {
			return err
		}

		return nil
	}

	// DISCONNECT <peer-id>
	if action == JACKIE_DISCONNECT {
		peer.UnsetNeighbor(args[0])

		log.Println("Disconnected to", args[0])

		return nil
	}

	// DOWNLOAD_BLOCKCHAIN <peer-id>
	if action == JACKIE_DOWNLOAD_BLOACKCHAIN {
		bs, err := json.Marshal(chain)
		if err != nil {
			return err
		}

		log.Println("Node", args[0], "requested blockchain download with size equals", len(bs), "bytes")

		schain := base64.StdEncoding.EncodeToString(bs)

		if err := DownloadBlockchainOK(peer.GetID(), schain, peer.GetNeighborByID(args[0])); err != nil {
			return err
		}
		return nil
	}

	// DOWNLOAD_BLOCKCHAIN_OK <peer-id> <chain-b64-encoded>
	if action == JACKIE_DOWNLOAD_BLOACKCHAIN_OK {
		raw := strings.Trim(args[1], "\x00")

		bs, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return err
		}

		cchain := new(blockchain.Chain)
		if err = json.NewDecoder(bytes.NewBuffer(bs)).Decode(chain); err != nil {
			return err
		}

		chain = cchain

		log.Println("Blockchain downloaded successful with size equals", len(bs), "bytes")
		return nil
	}

	// TX_APPROBATION <peer-id> <tx-encoded-b64>
	if action == JACKIE_TX_APPROBATION {
		raw := strings.Trim(args[1], "\x00")

		bs, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return err
		}

		tx := blockchain.Transaction{}
		if err := json.NewDecoder(bytes.NewBuffer(bs)).Decode(&tx); err != nil {
			return err
		}

		chain.AddPendingTransaction(tx)

		log.Println("Tx", tx.CalculateHash(), "received from peer", args[0])

		return nil
	}

	return errors.New("Invalid jackie action")
}

func HTTPHandler(peer Peer, chain *blockchain.Chain, upat time.Time, port string, conn net.Conn, req *http.Request) error {
	defer conn.Close()

	if req.Method == http.MethodGet {
		if req.URL.Path == "/transactions" {
			return ListTxsHTTPHandler(*chain, conn, req)
		}

		if req.URL.Path == "/blocks" {
			return ListBlocksHTTPHandler(*chain, conn, req)
		}

		if req.URL.Path == "/info" {
			return GetPeerInfoHTTPHandler(upat, port, peer, conn, req)
		}

		rg := regexp.MustCompile(`(/wallets/).+`)
		if matched := rg.MatchString(req.URL.Path); matched {
			parts := strings.Split(req.URL.Path, "/")
			return GetWalletBySeedHTTPHandler(parts[2], conn, req)
		}

		rg = regexp.MustCompile(`(/balance/).+`)
		if matched := rg.MatchString(req.URL.Path); matched {
			parts := strings.Split(req.URL.Path, "/")
			return GetBalanceByWalletAddress(parts[2], *chain, conn, req)
		}
	}

	if req.Method == http.MethodPost {
		if req.URL.Path == "/transactions" {
			return CreateTxHTTPHandler(peer, chain, conn, req)
		}

		if req.URL.Path == "/wallets" {
			return CreateWalletHTTPHandler(conn, req)
		}
	}

	return httputil.ResponseNotFound(req, conn)
}

func NewService(port string) *Service {
	return &Service{
		id:           uuid.NewString(),
		neighborhood: make(map[string]string),
	}
}
