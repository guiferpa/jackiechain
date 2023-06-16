package blockchain

type TransactionInput struct {
	Signature    []byte `json:"signature,omitempty"`
	TxOutputHash string `json:",omitempty"`
	Sender       string `json:"sender"`
	Amount       int    `json:"amount"`
}
