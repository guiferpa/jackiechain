package blockchain

import (
	"crypto/ed25519"

	"github.com/mr-tron/base58"
)

func GetUTXOsByWalletAddress(chain Chain, encseed string) ([]TransactionOutput, error) {
	seed, err := base58.Decode(encseed)
	if err != nil {
		return nil, err
	}

	priv := ed25519.NewKeyFromSeed(seed)

	addr := base58.Encode(priv.Public().(ed25519.PublicKey))

	txs := chain.UTXO[addr]

	return txs, nil
}
