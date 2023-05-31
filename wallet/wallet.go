package wallet

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/mr-tron/base58"
)

type Wallet struct {
	Name       string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func (w *Wallet) GetAddress() string {
	return base58.Encode(w.PublicKey)
}

func NewWallet(name string) (*Wallet, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Name:       name,
		PrivateKey: priv,
		PublicKey:  priv.Public().(ed25519.PublicKey),
	}, nil
}
