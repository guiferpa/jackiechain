package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/guiferpa/jackiechain/cmd/agent/actions"
	"github.com/guiferpa/jackiechain/logger"
	protogreeter "github.com/guiferpa/jackiechain/proto/greeter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func printPrompt() {
	fmt.Print("jackie > ")
}

func main() {
	serverHost := flag.String("server-host", "0.0.0.0", "server host")
	serverPort := flag.Int("server-port", 9000, "server port")

	flag.Parse()

	agentID := uuid.New().String()

	addr := fmt.Sprintf("%s:%v", *serverHost, *serverPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	greeter := protogreeter.NewGreeterClient(conn)

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

		if strings.ToLower(act) == actions.GreeterPing {
			resp, err := greeter.ReachOut(context.Background(), &protogreeter.PingRequest{
				Aid: agentID,
			})
			if err != nil {
				logger.Red(err.Error())
				return
			}
			logger.Yellow(fmt.Sprintf("Pong from peer %s", resp.Pid))
		}

		printPrompt()
	}
}
