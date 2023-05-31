package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Hash         string       `json:"hash"`
	PreviousHash string       `json:"previous_hash"`
	Nonce        int64        `json:"nonce"`
	Timestamp    time.Time    `json:"timestamp"`
	Transactions Transactions `json:"transactions"`
}

func (b *Block) CalculateHash() {
	payload := fmt.Sprintf("%s::%v::%s:%v", b.PreviousHash, b.Timestamp.UnixMilli(), b.Transactions.String(), b.Nonce)
	h := sha256.New()
	h.Write([]byte(payload))
	b.Hash = hex.EncodeToString(h.Sum(nil))
}

func NewBlock(transactions Transactions) *Block {
	return &Block{
		Timestamp:    time.Now(),
		Transactions: transactions,
	}
}
