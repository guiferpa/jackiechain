package transaction

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/mr-tron/base58"
)

type Tx struct {
	Sender    string     `json:"sender"`
	TxOuts    TxOutSlice `json:"tx_outs"`
	Signature []byte     `json:"signature"`
	Timestamp int64      `json:"timestamp"`
}

func (tx Tx) Bytes() ([]byte, error) {
	bs, err := json.Marshal(&tx)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

type TxSlice []Tx

func (txs TxSlice) GenerateTxHashes() ([]string, error) {
	hs := make([]string, 0)
	for _, tx := range txs {
		h, err := GenerateTxHash(tx)
		if err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}

type TxMap map[string]Tx

func GenerateTxHash(tx Tx) (string, error) {
	bs, err := tx.Bytes()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(bs)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func SignTx(tx Tx, privkey ed25519.PrivateKey) ([]byte, error) {
	s256 := sha256.New()
	h, err := GenerateTxHash(tx)
	if err != nil {
		return nil, err
	}
	s256.Write([]byte(h))
	return ed25519.Sign(privkey, s256.Sum(nil)), nil
}

func TxHasValidSignature(tx Tx) (bool, error) {
	b, err := base58.Decode(tx.Sender)
	if err != nil {
		return false, err
	}
	h, err := GenerateTxHash(tx)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(b, []byte(h), tx.Signature[:ed25519.SignatureSize]), nil
}
