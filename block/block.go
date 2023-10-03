package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/guiferpa/jackiechain/transaction"
)

type BlockHeader struct {
	Version            string `json:"version"`
	MerkleTreeRootHash string `json:"merkle_tree_root_hash"`
	Nonce              int    `json:"nonce"`
	Timestamp          int64  `json:"timestamp"`
	PreviousBlockHash  string `json:"previous_block_hash"`
}

func (bh BlockHeader) Bytes() ([]byte, error) {
	bs, err := json.Marshal(&bh)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

type Block struct {
	Header       BlockHeader
	Transactions transaction.TxMap
}

type BlockMap map[string]Block

func GenerateBlockHash(b Block) (string, error) {
	h := sha256.New()
	bs, err := b.Header.Bytes()
	if err != nil {
		return "", err
	}
	h.Write(bs)
	return hex.EncodeToString(h.Sum(nil)), nil
}
