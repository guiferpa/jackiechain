package blockchain

type Chain struct {
	Blocks              []Block
	Transactions        Transactions
	PendingTransactions Transactions
	MiningDifficulty    int
	MiningReward        int
	UTXO                map[string][]TransactionOutput
}

func (c *Chain) AddPendingTransaction(tx Transaction) error {
	ptxs := c.PendingTransactions
	c.PendingTransactions = append(ptxs, tx)
	return nil
}

func (c *Chain) MineBlock(miner string) string {
	ptxs := c.PendingTransactions

	previousBlock := c.GetLatestBlockFromChain()
	candidateBlock := NewBlock(CalculateBlockHash(previousBlock), ptxs)
	candidateBlock.Mine(c.MiningDifficulty, miner)

	c.Transactions = append(c.Transactions, ptxs...)

	cbtx := NewCoinbaseTransaction(CoinbaseTransactionOptions{
		Outputs: []TransactionOutput{
			{Receiver: miner, Amount: c.MiningReward},
		},
	})
	c.PendingTransactions = []Transaction{*cbtx}

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
	genesisBlock := Block{}

	return &Chain{
		MiningDifficulty:    opts.MiningDifficulty,
		MiningReward:        opts.MiningReward,
		PendingTransactions: make(Transactions, 0),
		Blocks:              []Block{genesisBlock},
	}
}
