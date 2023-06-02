package main

import (
	"fmt"

	"github.com/guiferpa/jackchain/wallet"
	"github.com/spf13/cobra"
)

var txCmd *cobra.Command

type CreateTransactionHTTPRequestBody struct {
	From   string // wallet address
	To     string // wallet address
	Amount int64
	Key    string // private key
	Port   string // jackchain api port
}

func init() {
	txCmd = &cobra.Command{
		Use:   "tx",
		Short: "Create assigned transaction",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			w, err := wallet.ParseWallet()
			if err != nil {
				panic(err)
			}

			fmt.Println(w.GetAddress())

			/*
				from := args[0]
				to := args[1]
				amount := args[2]
				key := args[3]
				pamount, err := strconv.ParseInt(amount, 64)
				if err != nil {
					panic(err)
				}

				blockchain.NewSignedTransaction(blockchain.TransactionOptions{

				})

				body := CreateTransactionHTTPRequestBody{
					From:   from,
					To:     to,
					Amount: pamount,
				}

				port := args[4]
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%s/transactions/sign"))
			*/
		},
	}
}
