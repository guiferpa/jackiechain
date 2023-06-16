package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mr-tron/base58"

	"github.com/guiferpa/jackiechain/wallet"
)

type Sender struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

type Receiver struct {
	PublicKey ed25519.PublicKey
}

type Transactions []Transaction

func (ts Transactions) String() string {
	b := bytes.NewBuffer([]byte(""))
	if err := json.NewEncoder(b).Encode(ts); err != nil {
		panic(err)
	}
	return b.String()
}

type Transaction struct {
	Signature []byte              `json:"signature"`
	Sender    *wallet.Wallet      `json:"-"`
	Inputs    []TransactionInput  `json:"inputs"`
	Outputs   []TransactionOutput `json:"outputs"`
	Timestamp int64               `json:"timestamp"`
}

func (t *Transaction) CalculateHash() string {
	payload := fmt.Sprintf("%s::%s::%v::%v", t.Timestamp)
	h := sha256.New()
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (t *Transaction) Sign(privKey ed25519.PrivateKey) {
	h := t.CalculateHash()
	s := sha256.New()
	s.Write([]byte(h))
	signature := ed25519.Sign(privKey, []byte(h))
	t.Signature = signature
}

func (t *Transaction) HasValidSignature() (bool, error) {
	if len(t.Inputs) == 0 {
		return true, nil
	}

	b, err := base58.Decode(t.Sender.GetAddress())
	if err != nil {
		return false, err
	}

	h := t.CalculateHash()

	return ed25519.Verify(b, []byte(h), t.Signature[:ed25519.SignatureSize]), nil
}

type TransactionOptions struct {
	Sender  *wallet.Wallet
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

func NewTransaction(opts TransactionOptions) *Transaction {
	return &Transaction{
		Sender:    opts.Sender,
		Inputs:    opts.Inputs,
		Outputs:   opts.Outputs,
		Timestamp: time.Now().UnixNano(),
	}
}

func NewSignedTransaction(opts TransactionOptions) *Transaction {
	t := NewTransaction(opts)
	t.Sign(opts.Sender.PrivateKey)
	return t
}

type CoinbaseTransactionOptions struct {
	Outputs []TransactionOutput
}

func NewCoinbaseTransaction(opts CoinbaseTransactionOptions) *Transaction {
	return NewTransaction(TransactionOptions{
		Inputs:  make([]TransactionInput, 0),
		Outputs: opts.Outputs,
	})
}
