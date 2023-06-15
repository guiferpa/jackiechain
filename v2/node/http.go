package node

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/httputil"
)

type CreateTxRequestBody struct {
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	PrivateSeed string `json:"private_seed"`
	Amount      int    `json:"amount"`
}

func CreateTxHandler(chain blockchain.Chain, conn net.Conn, req *http.Request) error {
	body := CreateTxRequestBody{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	utxos, err := blockchain.GetUTXOsByWalletAddress(chain, body.PrivateSeed)
	if err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	balance := 0
	inputs := make([]blockchain.TransactionInput, 0)
	for _, tx := range utxos {
		balance += tx.Amount
		inputs = append(inputs, *tx.ToInput())
	}

	if balance < body.Amount {
		return httputil.ResponseBadRequest(req, conn, "insufficient funds")
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(inputs); err != nil {
		return httputil.ResponseBadRequest(req, conn, err.Error())
	}

	return httputil.Response(req, conn, http.StatusOK, buf)
}
