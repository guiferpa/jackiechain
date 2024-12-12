package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/guiferpa/jackiechain/agent"
	"github.com/guiferpa/jackiechain/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverHost := flag.String("server-host", "0.0.0.0", "server host")
	serverPort := flag.Int("server-port", 9000, "server port")

	flag.Parse()

	addr := fmt.Sprintf("%s:%v", *serverHost, *serverPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Red(err.Error())
		return
	}

	a := agent.New(conn)

	scanner := bufio.NewScanner(os.Stdin)
	if err := a.ExecPrompt(scanner); err != nil {
		logger.Red(err.Error())
    return
	}
}
