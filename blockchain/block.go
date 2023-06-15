package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type BlockHeader struct {
	Version           string    `json:"version"`
	Nonce             int       `json:"nonce"`
	Timestamp         time.Time `json:"timestamp"`
	PreviousBlockHash string    `json:"previous_block_hash"`
}

func (bh BlockHeader) Buffer() *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(
		"%s::%d::%d::%s",
		bh.PreviousBlockHash,
		bh.Timestamp.UnixMilli(),
		bh.Nonce,
		bh.Version,
	))
}

type Block struct {
	Header       BlockHeader  `json:"header"`
	Transactions Transactions `json:"transactions"`
}

func (b *Block) Mine(dif int, miner string) {
	got := strings.Repeat("0", dif)

	hash := CalculateBlockHash(*b)

	for got != hash[0:dif] {
		b.Header.Nonce++
		hash = CalculateBlockHash(*b)
	}
}

func NewBlock(pbh string, transactions Transactions) *Block {
	return &Block{
		Header: BlockHeader{
			PreviousBlockHash: pbh,
			Timestamp:         time.Now(),
		},
		Transactions: transactions,
	}
}

func CalculateBlockHash(block Block) string {
	hasher := sha256.New()
	hasher.Write(block.Header.Buffer().Bytes())
	return hex.EncodeToString(hasher.Sum(nil))
}
