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

type CreateTxHTTPRequestBody struct {
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	PrivateSeed string `json:"private_seed"`
	Amount      int    `json:"amount"`
}

func CreateTxHTTPHandler(peer Peer, chain *blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	body := CreateTxHTTPRequestBody{}
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

	inputs, err := blockchain.GetUTXOsByWalletAddress(*chain, w.GetAddress())
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

	for _, neighbor := range peer.GetNeighborhood() {
		// Ignore net failure
		TxApprobationRequest(peer.GetID(), *tx, neighbor)
	}

	return httputil.Response(req, conn, http.StatusNoContent, nil)
}

type ListTxsHTTPResponseBody blockchain.Transactions

func ListTxsHTTPHandler(chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	txs := make(ListTxsHTTPResponseBody, len(chain.Transactions))

	for i, tx := range chain.Transactions {
		txs[cap(txs)-(i+1)] = tx
	}

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

func ListBlocksHTTPHandler(chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	blocks := make(ListBlockResponseBody, len(chain.Blocks))

	for i, block := range chain.Blocks {
		blocks[cap(blocks)-(i+1)] = ListBlockInfo{
			Block: block,
			Hash:  blockchain.CalculateBlockHash(block),
		}
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

type GetPeerInfoHTTPResponseBody struct {
	ID          string        `json:"id"`
	Uptime      time.Duration `json:"uptime"`
	Peers       []PeerInfo    `json:"peers"`
	NodePort    string        `json:"node_port"`
	MiningClock string        `json:"mining_clock"`
}

func GetPeerInfoHTTPHandler(upat time.Time, port string, peer Peer, conn net.Conn, req *http.Request) error {
	uptime := time.Now().Sub(upat)

	peers := make([]PeerInfo, 0)
	for id, addr := range peer.GetNeighborhood() {
		peers = append(peers, PeerInfo{ID: id, Addr: addr})
	}

	body := GetPeerInfoHTTPResponseBody{
		ID:          peer.GetID(),
		NodePort:    port,
		Peers:       peers,
		Uptime:      time.Duration(uptime / time.Second),
		MiningClock: "",
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}

type CreateWalletHTTPResponseBody struct {
	Address     string `json:"address"`
	PrivateSeed string `json:"private_seed"`
}

func CreateWalletHTTPHandler(conn net.Conn, req *http.Request) error {
	w, err := wallet.NewWallet()
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	body := CreateWalletHTTPResponseBody{
		Address:     w.GetAddress(),
		PrivateSeed: w.GetPrivateSeed(),
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusCreated, buf)
}

type GetWalletHTTPResponseBody CreateWalletHTTPResponseBody

func GetWalletBySeedHTTPHandler(seed string, conn net.Conn, req *http.Request) error {
	w, err := wallet.ParseWallet(seed)
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	body := GetWalletHTTPResponseBody{
		Address:     w.GetAddress(),
		PrivateSeed: w.GetPrivateSeed(),
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}

type GetBalanceByWalletAddressHTTPResponseBody struct {
	Date    string `json:"date"`
	Balance int    `json:"balance"`
}

func GetBalanceByWalletAddress(addr string, chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	txs, err := blockchain.GetUTXOsByWalletAddress(chain, addr)
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	balance := 0
	for _, tx := range txs {
		balance += tx.Amount
	}

	body := GetBalanceByWalletAddressHTTPResponseBody{
		Balance: balance,
		Date:    time.Now().Format(time.RFC3339),
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}
