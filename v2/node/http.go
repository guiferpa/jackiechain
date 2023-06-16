package node

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/httputil"
	"github.com/guiferpa/jackiechain/wallet"
)

type CreateTxRequestBody struct {
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	PrivateSeed string `json:"private_seed"`
	Amount      int    `json:"amount"`
}

func CreateTxHandler(chain *blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	body := CreateTxRequestBody{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	w, err := wallet.ParseWallet(body.PrivateSeed)
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	if w.GetAddress() != body.Sender {
		return httputil.ResponseBadRequest(req, conn, "seed is wrong for your wallet")
	}

	inputs, err := blockchain.GetUTXOsByWalletAddress(*chain, *w)
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	balance := 0
	for _, tx := range inputs {
		balance += tx.Amount
	}

	if balance < body.Amount {
		return httputil.ResponseBadRequest(req, conn, "insufficient funds")
	}

	tx := blockchain.NewSignedTransaction(blockchain.TransactionOptions{
		Sender: w,
		Inputs: inputs,
		Outputs: []blockchain.TransactionOutput{
			*blockchain.NewTransactionOutput(blockchain.TransactionOutputOptions{
				Receiver: body.Receiver, Amount: body.Amount,
			}),
			*blockchain.NewTransactionOutput(blockchain.TransactionOutputOptions{
				Receiver: w.GetAddress(), Amount: balance - body.Amount,
			}),
		},
	})

	chain.AddPendingTransaction(*tx)

	return httputil.Response(req, conn, http.StatusNoContent, nil)
}

func ListTxsHandler(chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	txs := chain.Transactions

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(txs); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}

type ListBlockInfo struct {
	blockchain.Block
	Hash string `json:"hash"`
}

type ListBlockResponseBody []ListBlockInfo

func ListBlocksHandler(chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	blocks := make(ListBlockResponseBody, 0)

	for _, block := range chain.Blocks {
		blocks = append(blocks, ListBlockInfo{
			Block: block,
			Hash:  blockchain.CalculateBlockHash(block),
		})
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(blocks); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}

type PeerInfo struct {
	ID   string `json:"id"`
	Addr string `json:"address"`
}

type GetPeerInfoResponseBody struct {
	ID       string        `json:"id"`
	Uptime   time.Duration `json:"uptime"`
	Peers    []PeerInfo    `json:"peers"`
	NodePort string        `json:"node_port"`
}

func GetPeerInfoHandler(upat time.Time, port string, peer Peer, conn net.Conn, req *http.Request) error {
	uptime := time.Now().Sub(upat)

	peers := make([]PeerInfo, 0)
	for id, addr := range peer.GetNeighborhood() {
		peers = append(peers, PeerInfo{ID: id, Addr: addr})
	}

	body := GetPeerInfoResponseBody{
		ID:       peer.GetID(),
		NodePort: port,
		Peers:    peers,
		Uptime:   time.Duration(uptime / time.Second),
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}
