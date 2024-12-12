package agent

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/guiferpa/jackiechain/agent/actions"
	"github.com/guiferpa/jackiechain/logger"
	"github.com/guiferpa/jackiechain/proto/greeter"
	"google.golang.org/grpc"
)

const PROMPT_LABEL = "jackie > "

type ID string

type protoClients struct {
	Greeter greeter.GreeterClient
}

type Agent struct {
	ID           ID
	protoClients protoClients
}

func (a *Agent) ExecPrompt(scanner *bufio.Scanner) error {
	fmt.Print(PROMPT_LABEL)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		args := strings.Fields(scanner.Text())

		if len(args) == 0 {
			fmt.Print(PROMPT_LABEL)
			continue
		}

		act := args[0]

		if strings.ToLower(act) == actions.GreeterPing {
			resp, err := a.protoClients.Greeter.ReachOut(context.Background(), &greeter.PingRequest{
				Aid: string(a.ID),
			})
			if err != nil {
				return err
			}
			logger.Yellow(fmt.Sprintf("Pong from peer %s", resp.Pid))
		}

		fmt.Print(PROMPT_LABEL)
	}

	return nil
}

func New(conn grpc.ClientConnInterface) *Agent {
	return &Agent{
		ID: ID(uuid.NewString()),
		protoClients: protoClients{
			Greeter: greeter.NewGreeterClient(conn),
		},
	}
}
