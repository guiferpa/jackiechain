package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd *cobra.Command

func init() {
	listCmd = &cobra.Command{
		Use:   "list [resource]",
		Short: "List some resource",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Error: must also specify a resource")
		},
	}
}
