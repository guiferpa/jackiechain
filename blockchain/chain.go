package blockchain

import (
	"fmt"
	"strings"

	"github.com/guiferpa/jackchain/wallet"
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

func (c *Chain) AddTransaction(sender wallet.Wallet, receiver string, amount int64) {
	transaction := NewSignedTransaction(TransactionOptions{
		Sender:       sender,
		ReceiverAddr: receiver,
		Amount:       amount,
	})

	has, err := transaction.HasValidSignature()
	if err != nil {
		panic(err)
	}

	if !has {
		fmt.Printf("invalid signature (%s) for transaction (%s)\n", transaction.Signature, transaction.CalculateHash())
		return
	}

	c.PendingTransactions = append(c.PendingTransactions, *transaction)
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
