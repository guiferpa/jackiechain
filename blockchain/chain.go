package blockchain

import (
	"errors"
	"fmt"
	"strings"
)

type Chain struct {
	Blocks              []Block      `json:"blocks"`
	PendingTransactions Transactions `json:"pending_transactions"`
	MiningDifficulty    int          `json:"mining_difficulty"`
	MiningReward        int          `json:"mining_reward"`
}

func (c *Chain) MinePendingTransactions() {
	block := NewBlock(c.PendingTransactions)

	got := strings.Repeat("0", c.MiningDifficulty)

	block.CalculateHash()

	for got != block.Hash[0:c.MiningDifficulty] {
		block.Nonce++
		block.CalculateHash()
	}

	c.Blocks = append(c.Blocks, *block)

	c.PendingTransactions = make(Transactions, 0)
}

func (c *Chain) AddTransaction(tx *Transaction) error {
	has, err := tx.HasValidSignature()
	if err != nil {
		return err
	}

	if !has {
		return errors.New(fmt.Sprintf("invalid signature (%s) for transaction (%s)\n", tx.Signature, tx.CalculateHash()))
	}

	c.PendingTransactions = append(c.PendingTransactions, *tx)
	return nil
}

type ChainOptions struct {
	MiningDifficulty    int
	MiningReward        int
	PendingTransactions Transactions
}

func NewChain(opts ChainOptions) *Chain {
	return &Chain{
		MiningDifficulty:    opts.MiningDifficulty,
		MiningReward:        opts.MiningReward,
		PendingTransactions: opts.PendingTransactions,
		Blocks:              make([]Block, 0),
	}
}
