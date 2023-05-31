// https://dev.to/nheindev/building-a-blockchain-in-go-pt-v-wallets-12na

package main

import "github.com/spf13/cobra"

func main() {
	rootCmd := &cobra.Command{}

	createCmd.AddCommand(createWalletCmd)

	rootCmd.AddCommand(createCmd)

	rootCmd.Execute()
}
