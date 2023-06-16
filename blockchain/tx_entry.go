package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

type TransactionEntry interface {
	Buffer() *bytes.Buffer
}

func CalculateTxEntryHash(tx TransactionEntry) string {
	hasher := sha256.New()
	hasher.Write(tx.Buffer().Bytes())
	return hex.EncodeToString(hasher.Sum(nil))
}
