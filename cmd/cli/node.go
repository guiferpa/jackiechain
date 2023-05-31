package main

import (
	"fmt"

	"github.com/guiferpa/jackchain/api"
	"github.com/guiferpa/jackchain/blockchain"
	"github.com/guiferpa/jackchain/wallet"
	"github.com/spf13/cobra"
)

var (
	createNodeCmd *cobra.Command
)

func init() {
	createNodeCmd = &cobra.Command{
		Use:   "node [id] [port]",
		Short: "Create a new node",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			port := args[1]

			sender, err := wallet.NewWallet("sender")
			if err != nil {
				panic(err)
			}

			receiver, err := wallet.NewWallet("receiver")
			if err != nil {
				panic(err)
			}

			chain := blockchain.NewChain(blockchain.ChainOptions{
				MiningDifficulty:    2,
				MiningReward:        100,
				PendingTransactions: nil,
			})

			chain.AddTransaction(*sender, receiver.GetAddress(), 100)

			chain.MinePendingTransactions()

			fmt.Printf("Node (%s) is running at port: %v\n", id, port)

			api.Run(*chain, port)
		},
	}
}
