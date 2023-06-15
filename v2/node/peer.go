package node

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

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
	if _, exists := s.neighborhood[id]; exists {
		return errors.New("duplicated peer")
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

func JackieHandler(peer Peer, chain *blockchain.Chain, port, action string, args []string) error {
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
			if err.Error() == "duplicated peer" {
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
			if err.Error() == "duplicated peer" {
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

		mu.Lock()

		cchain := new(blockchain.Chain)
		err = json.NewDecoder(bytes.NewBuffer(bs)).Decode(chain)
		chain = cchain

		mu.Unlock()

		if err != nil {
			return err
		}

		log.Println("Blockchain downloaded successful with size equals", len(bs), "bytes")
		return nil
	}

	return errors.New("Invalid jackie action")
}

func HTTPHandler(chain *blockchain.Chain, conn net.Conn, req *http.Request) error {
	defer conn.Close()

	if req.Method == http.MethodPost {
		if req.URL.Path == "/tx" {
			return CreateTxHandler(*chain, conn, req)
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
