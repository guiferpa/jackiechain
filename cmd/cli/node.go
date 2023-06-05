package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/guiferpa/jackchain/blockchain"
	"github.com/guiferpa/jackchain/net"
	"github.com/spf13/cobra"
)

var (
	createNodeCmd *cobra.Command
)

func init() {
	createNodeCmd = &cobra.Command{
		Use:   "node [addr]",
		Short: "Create a new node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.ParseFlags(args); err != nil {
				panic(err)
			}

			nested, err := cmd.Flags().GetString("join")
			if err != nil {
				panic(err)
			}

			addr := args[0]
			chain := blockchain.NewChain(blockchain.ChainOptions{})

			node, err := net.NewNode(addr, chain)
			if err != nil {
				panic(err)
			}

			doneCh := make(chan bool, 1)
			errCh := make(chan error, 1)
			sigCh := make(chan os.Signal, 1)

			signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

			go node.Listen(doneCh, errCh)

			if nested != "" {
				go node.Join(nested, doneCh, errCh)
			}

			for {
				select {
				case err := <-errCh:
					fmt.Println(err)

				case <-sigCh:
					if err := node.Disconnect(nested); err != nil {
						panic(err)
					}

					os.Exit(0)

				default:
				}
			}
		},
	}

	createNodeCmd.PersistentFlags().String("join", "", "host:port")
}
