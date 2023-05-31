package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd *cobra.Command

func init() {
	createCmd = &cobra.Command{
		Use:   "create [resource]",
		Short: "Create some resource",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Error: must also specify a resource")
		},
	}
}
