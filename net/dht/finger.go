package dht

import (
	"math/big"
)

type Finger struct {
	Id   []byte
	Node Node
}

func NewFinger() *Finger {
	return nil
}

// Computes the offset by (n + 2^i) mod (2^m)
func CalculateFingerId(n []byte, i, m int) []byte {
	// Convert the ID to a bigint
	idInt := (&big.Int{}).SetBytes(n)

	// Get the offset
	two := big.NewInt(2)
	offset := big.Int{}
	offset.Exp(two, big.NewInt(int64(i)), nil)

	// Sum
	sum := big.Int{}
	sum.Add(idInt, &offset)

	// Get the ceiling
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(m)), nil)

	// Apply the mod
	idInt.Mod(&sum, &ceil)
	// Add together
	return idInt.Bytes()
}
