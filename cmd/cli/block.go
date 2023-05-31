package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	listBlockCmd *cobra.Command
)

func init() {
	listBlockCmd = &cobra.Command{
		Use:   "block [port]",
		Short: "List blocks",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port := args[0]
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%s/chain", port), nil)
			if err != nil {
				panic(err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()

			var body map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				panic(body)
			}

			fmt.Println(body["blocks"])
		},
	}
}
