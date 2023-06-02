package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mr-tron/base58"
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func (w *Wallet) GetAddress() string {
	return base58.Encode(w.PublicKey)
}

func (w *Wallet) ExportPrivateKey() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dst := make([]byte, hex.EncodedLen(len(w.PrivateKey.Seed())))
	hex.Encode(dst, w.PrivateKey.Seed())

	if err := ioutil.WriteFile(fmt.Sprintf("%s/key.pem", pwd), dst, 0600); err != nil {
		return err
	}

	return nil
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

func ParseWallet() (*Wallet, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	bs, err := ioutil.ReadFile(fmt.Sprintf("%s/key.pem", pwd))
	if err != nil {
		return nil, err
	}

	dst := make([]byte, hex.DecodedLen(len(bs)))
	hex.Decode(dst, bs)

	priv := ed25519.NewKeyFromSeed(dst)

	w := &Wallet{
		PrivateKey: priv,
		PublicKey:  priv.Public().(ed25519.PublicKey),
	}

	return w, nil
}
