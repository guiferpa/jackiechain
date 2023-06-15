package blockchain

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type TransactionInput struct {
	Signature []byte
	Sender    string
	Amount    int
}

func (ti TransactionInput) HasValidSignature(priv ed25519.PrivateKey) error {
	pub := priv.Public().(rsa.PublicKey)
	if _, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &pub, []byte("testing"), nil); err != nil {
		return err
	}

	return nil
}
