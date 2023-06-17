package node

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/logger"
)

const MAX_CHUNK_SIZE = 1024

func Listen(port string, verbose bool, peer Peer, chain *blockchain.Chain) {
	upat := time.Now()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		bs := make([]byte, MAX_CHUNK_SIZE)
		_, err = conn.Read(bs)
		if err == io.EOF {
			continue
		}
		if err != nil {
			logger.Red(err)
			continue
		}

		buf := bytes.NewBuffer(bs)

		if verbose {
			logger.Yellow(buf.String())
		}

		reader := bufio.NewReader(bytes.NewBuffer(buf.Bytes()))

		if req, err := http.ReadRequest(reader); err == nil {
			if err := HTTPHandler(peer, chain, upat, port, conn, req); err != nil {
				logger.Red(err)
			}
		} else {
			line := strings.Fields(buf.String())
			if len(line) < 3 {
				logger.Red("invalid protocol")
			}
			if err := JackieHandler(peer, chain, upat, port, line[1], line[2:]); err != nil {
				logger.Red(err)
			}
		}
	}
}

func MineNewBlock(wallet string, peer Peer, chain *blockchain.Chain, miningTicker *time.Ticker) {
	for range miningTicker.C {
		mu.Lock()
		block := chain.MineBlock(wallet)
		log.Println("Block", blockchain.CalculateBlockHash(block), "mined")

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(&block); err != nil {
			logger.Red(err.Error())
			continue
		}

		encblock := base64.StdEncoding.EncodeToString(buf.Bytes())

		for _, neighbor := range peer.GetNeighborhood() {
			if err := BlockApprobationRequest(peer.GetID(), encblock, neighbor); err != nil {
				logger.Red(err.Error())
			}
		}
		mu.Unlock()
	}
}

func TerminatePeer(peer Peer) {
	mu.Lock()
	defer mu.Unlock()

	for _, neighbor := range peer.GetNeighborhood() {
		PeerDisconnect(peer.GetID(), neighbor)
	}
}
