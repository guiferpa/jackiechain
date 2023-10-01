package main

import (
	"fmt"
	"time"

	"github.com/guiferpa/jackiechain/block"
	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/logger"
	"github.com/guiferpa/jackiechain/transaction"
)

func main() {
	bc := &blockchain.Blockchain{
		Blocks:           make(block.BlockMap),
		Txs:              make(transaction.TxMap),
		PendingTxs:       make(transaction.TxMap),
		MiningDifficulty: 4,
		UTxOs:            make(transaction.UTxOMap),
		GenesisBlock:     nil,
		LatestBlock:      nil,
	}

	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ticker.C:
			bh, err := blockchain.BuildBlock(bc)
			if err != nil {
				logger.Red(err.Error())
				continue
			}
			logger.Magenta(fmt.Sprintf("Block %s was built", bh))

		}
	}
}
