package blockchain

import "fmt"

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

func (c *Chain) MineBlock(miner string) string {
	ptxs := c.PendingTransactions

	previousBlock := c.GetLatestBlockFromChain()
	candidateBlock := NewBlock(CalculateBlockHash(previousBlock), ptxs)
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

	return CalculateBlockHash(*candidateBlock)
}

func (c *Chain) GetLatestBlockFromChain() Block {
	return c.Blocks[len(c.Blocks)-1]
}

type ChainOptions struct {
	MiningDifficulty int
	MiningReward     int
}

func NewChain(opts ChainOptions) *Chain {
	genesisBlock := NewBlock("", []Transaction{})

	return &Chain{
		MiningDifficulty:    opts.MiningDifficulty,
		MiningReward:        opts.MiningReward,
		PendingTransactions: make(Transactions, 0),
		Transactions:        make(Transactions, 0),
		UTXO:                make(map[string]TransactionOutput),
		Blocks:              []Block{*genesisBlock},
	}
}
