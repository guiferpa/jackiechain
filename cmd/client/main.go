package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/guiferpa/jackiechain/dist/proto"
	"github.com/guiferpa/jackiechain/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func printPrompt() {
	fmt.Print("jackie > ")
}

func main() {
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	client := proto.NewGreeterClient(conn)

	scanner := bufio.NewScanner(os.Stdin)

	printPrompt()

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			logger.Red(err.Error())
			break
		}
		args := strings.Fields(scanner.Text())

		if len(args) == 0 {
			printPrompt()
			continue
		}

		act := args[0]

		if strings.ToLower(act) == "ping" {
			resp, err := client.ReachOut(context.Background(), &proto.PingRequest{})
			if err != nil {
				logger.Red(err.Error())
				return
			}
			logger.Yellow(resp.Text)
		}

		printPrompt()
	}
}
