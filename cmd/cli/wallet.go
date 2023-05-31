package main

import (
	"fmt"

	"github.com/guiferpa/jackchain/wallet"
	"github.com/spf13/cobra"
)

var (
	createWalletCmd *cobra.Command
)

func init() {
	createWalletCmd = &cobra.Command{
		Use:   "wallet [name]",
		Short: "Create a wallet",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			w, err := wallet.NewWallet(name)
			if err != nil {
				fmt.Println("Error: ", err)
			}

			fmt.Println("Wallet created successful")
			fmt.Println("Address:", w.GetAddress())
		},
	}
}
