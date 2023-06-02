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
		Use:   "wallet",
		Short: "Create a wallet",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			w, err := wallet.NewWallet()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			fmt.Println("Wallet created successful")
			fmt.Println("Address:", w.GetAddress())

			w.ExportPrivateKey()
		},
	}
}
