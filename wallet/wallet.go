package wallet

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/mr-tron/base58"
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func (w *Wallet) GetAddress() string {
	return base58.Encode(w.PublicKey)
}

func (w *Wallet) GetPrivateSeed() string {
	return base58.Encode(w.PrivateKey.Seed())
}

func NewWallet() (*Wallet, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: priv,
		PublicKey:  priv.Public().(ed25519.PublicKey),
	}, nil
}

func ParseWallet(raw string) (*Wallet, error) {
	seed, err := base58.Decode(raw)
	if err != nil {
		return nil, err
	}

	priv := ed25519.NewKeyFromSeed(seed)

	w := &Wallet{
		PrivateKey: priv,
		PublicKey:  priv.Public().(ed25519.PublicKey),
	}

	return w, nil
}
