package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type UTxO struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	TxHash   string `json:"tx_hash"`
	Value    int64  `json:"value"`
}

func (utxo UTxO) Bytes() ([]byte, error) {
	bs, err := json.Marshal(&utxo)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func GenerateUTxOHash(utxo UTxO) (string, error) {
	bs, err := utxo.Bytes()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(bs)
	return hex.EncodeToString(h.Sum(nil)), nil
}

type UTxOSlice []UTxO

type UTxOMap map[string]UTxO
