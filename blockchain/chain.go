package blockchain

import (
	"fmt"
	"strings"
)

type Chain struct {
	Blocks              []Block
	Transactions        Transactions
	PendingTransactions Transactions
	MiningDifficulty    int
	MiningReward        int
	UTXO                map[string]TransactionOutput
}

func (c *Chain) AddPendingTransaction(tx Transaction) {
	for _, txin := range tx.Inputs {
		id := fmt.Sprintf("%s::%s", txin.TxOutputHash, txin.Sender)
		delete(c.UTXO, id)
	}

	for _, txout := range tx.Outputs {
		id := fmt.Sprintf("%s::%s", CalculateTxEntryHash(&txout), txout.Receiver)
		c.UTXO[id] = txout
	}

	ptxs := c.PendingTransactions
	c.PendingTransactions = append(ptxs, tx)
}

func (c *Chain) MineBlock(miner string) Block {
	ptxs := c.PendingTransactions

	previousBlockHash := c.GetLatestBlockHash()
	candidateBlock := NewBlock(previousBlockHash, ptxs)
	candidateBlock.Mine(c.MiningDifficulty, miner)

	c.Transactions = append(c.Transactions, ptxs...)
	c.PendingTransactions = []Transaction{}

	c.Blocks = append(c.Blocks, *candidateBlock)

	cbtx := NewCoinbaseTransaction(CoinbaseTransactionOptions{
		Outputs: []TransactionOutput{
			*NewTransactionOutput(TransactionOutputOptions{
				Receiver: miner,
				Amount:   c.MiningReward,
			}),
		},
	})
	c.AddPendingTransaction(*cbtx)

	return *candidateBlock
}

func (c *Chain) GetLatestBlockHash() string {
	if len(c.Blocks) == 0 {
		return strings.Repeat("0", 64)
	}

	return CalculateBlockHash(c.Blocks[len(c.Blocks)-1])
}

type ChainOptions struct {
	MiningDifficulty int
	MiningReward     int
}

func NewChain(opts ChainOptions) *Chain {
	return &Chain{
		MiningDifficulty:    opts.MiningDifficulty,
		MiningReward:        opts.MiningReward,
		PendingTransactions: make(Transactions, 0),
		Transactions:        make(Transactions, 0),
		UTXO:                make(map[string]TransactionOutput),
		Blocks:              []Block{},
	}
}
