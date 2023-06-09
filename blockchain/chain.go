package blockchain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guiferpa/jackiechain/wallet"
)

type Chain struct {
	Blocks              []Block      `json:"blocks"`
	PendingTransactions Transactions `json:"pending_transactions"`
	MiningDifficulty    int          `json:"mining_difficulty"`
	MiningReward        int          `json:"mining_reward"`
}

func (c *Chain) MinePendingTransactions(minerAddr string) string {
	block := NewBlock(c.PendingTransactions)

	got := strings.Repeat("0", c.MiningDifficulty)

	block.CalculateHash()

	for got != block.Hash[0:c.MiningDifficulty] {
		block.Nonce++
		block.CalculateHash()
	}

	if len(c.Blocks) > 0 {
		block.PreviousHash = c.Blocks[len(c.Blocks)-1].Hash
	}

	c.Blocks = append(c.Blocks, *block)

	rewardTx := NewTransaction(TransactionOptions{
		Sender:       wallet.Wallet{},
		ReceiverAddr: minerAddr,
		Amount:       int64(c.MiningReward),
	})
	c.PendingTransactions = Transactions{*rewardTx}

	return block.Hash
}

func (c *Chain) AddTransaction(tx *Transaction) error {
	has, err := tx.HasValidSignature()
	if err != nil {
		return err
	}

	if !has {
		return errors.New(fmt.Sprintf("invalid signature (%s) for transaction (%s)\n", tx.Signature, tx.CalculateHash()))
	}

	pt := c.PendingTransactions
	c.PendingTransactions = append(pt, *tx)

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
