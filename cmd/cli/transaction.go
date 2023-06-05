package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/guiferpa/jackchain/blockchain"
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
		Use:   "tx [to] [amount] [node]",
		Short: "Create assigned transaction",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			w, err := wallet.ParseWallet()
			if err != nil {
				panic(err)
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}

			tx := blockchain.NewSignedTransaction(blockchain.TransactionOptions{
				Sender:       *w,
				ReceiverAddr: args[0],
				Amount:       amount,
			})

			fmt.Println(tx.Signature)
			message := fmt.Sprintf("OK TX %s %s %d %s", tx.Sender, tx.Receiver, tx.Amount, tx.Signature)

			node := args[2]
			conn, err := net.Dial("tcp", node)
			if err != nil {
				panic(err)
			}

			b := bytes.NewBufferString(message)
			if _, err := conn.Write(b.Bytes()); err != nil {
				panic(err)
			}

			fmt.Println("Transaction", tx.CalculateHash(), "sent")
		},
	}
}
