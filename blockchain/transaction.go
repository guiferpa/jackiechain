package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

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
	Signature []byte `json:"signature"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Amount    int64  `json:"amount"`
}

func (t *Transaction) CalculateHash() string {
	payload := fmt.Sprintf("%s::%s::%v", t.Sender, t.Receiver, t.Amount)
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
	b, err := base58.Decode(t.Sender)
	if err != nil {
		return false, err
	}

	h := t.CalculateHash()

	return ed25519.Verify(b, []byte(h), t.Signature[:ed25519.SignatureSize]), nil
}

type TransactionOptions struct {
	Sender       wallet.Wallet
	ReceiverAddr string
	Amount       int64
}

func NewTransaction(opts TransactionOptions) *Transaction {
	return &Transaction{
		Sender:   opts.Sender.GetAddress(),
		Receiver: opts.ReceiverAddr,
		Amount:   opts.Amount,
	}
}

func NewSignedTransaction(opts TransactionOptions) *Transaction {
	t := NewTransaction(opts)
	t.Sign(opts.Sender.PrivateKey)
	return t
}
