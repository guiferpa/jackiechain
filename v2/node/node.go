package node

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/v2/logger"
)

const MAX_CHUNK_SIZE = 1024

func Listen(port string, verbose bool, peer Peer, chain *blockchain.Chain) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		bs := make([]byte, MAX_CHUNK_SIZE)
		_, err = conn.Read(bs)
		if err == io.EOF {
			continue
		}
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(bs)

		if verbose {
			logger.Yellow(buf.String())
		}

		reader := bufio.NewReader(bytes.NewBuffer(buf.Bytes()))

		if req, err := http.ReadRequest(reader); err == nil {
			if err := HTTPHandler(conn, req); err != nil {
				logger.Red(err)
			}
		} else {
			line := strings.Fields(buf.String())
			if len(line) < 3 {
				logger.Red("invalid protocol")
			}
			if err := JackieHandler(peer, chain, port, line[1], line[2:]); err != nil {
				logger.Red(err)
			}
		}
	}
}

func MineNewBlock(wallet string, chain *blockchain.Chain) {
	ticker := time.NewTicker(2 * time.Minute)

	for range ticker.C {
		mu.Lock()

		bhash := chain.MineBlock(wallet)

		mu.Unlock()

		log.Println("Block", bhash, "mined")
	}
}
