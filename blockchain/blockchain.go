package blockchain

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/guiferpa/jackiechain/block"
	"github.com/guiferpa/jackiechain/transaction"
)

type Blockchain struct {
	Blocks           block.BlockMap
	Txs              transaction.TxMap
	PendingTxs       transaction.TxMap
	MiningDifficulty int
	UTxOs            transaction.UTxOMap
	GenesisBlock     *block.Block
	LatestBlock      *block.Block
}

func MiningBlock(bc *Blockchain, b *block.Block) (string, error) {
	challenge := strings.Repeat("0", bc.MiningDifficulty)
	h, err := block.GenerateBlockHash(*b)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for challenge != h[0:bc.MiningDifficulty] {
		b.Header.Nonce++
		h, err = block.GenerateBlockHash(*b)
		if err != nil {
			return "", err
		}
	}
	return h, nil
}

func BuildBlock(bc *Blockchain) (string, error) {
	b := &block.Block{
		Header: block.BlockHeader{
			Version:   "1",
			Timestamp: time.Now().UnixMilli(),
		},
		Transactions: bc.PendingTxs,
	}
	h, err := MiningBlock(bc, b)
	if err != nil {
		return "", err
	}
	if bc.GenesisBlock == nil {
		b.Header.PreviousBlockHash = strings.Repeat("0", 64)
		bc.GenesisBlock = b
		bc.LatestBlock = b
	} else {
		pv, err := block.GenerateBlockHash(*bc.LatestBlock)
		if err != nil {
			return "", err
		}
		b.Header.PreviousBlockHash = pv
		bc.LatestBlock = b
	}
	bc.Blocks[h] = *b
	maps.Copy(bc.Txs, bc.PendingTxs)
	bc.PendingTxs = make(transaction.TxMap)
	return h, nil
}

func AddTx(bc *Blockchain, tx transaction.Tx) error {
	h, err := transaction.GenerateTxHash(tx)
	if err != nil {
		return err
	}
	has, err := transaction.TxHasValidSignature(tx)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("no valid tx signature")
	}
	bc.PendingTxs[h] = tx
	for _, txout := range tx.TxOuts {
		utxo := transaction.GenerateUTxOFromTxOut(txout)
		utxoh, err := transaction.GenerateUTxOHash(utxo)
		if err != nil {
			return err
		}
		bc.UTxOs[utxoh] = utxo
	}
	return nil
}
