package tcp

import (
	"bytes"
	"errors"
	"strings"
)

const (
	JACKIE_CONNECT    = "CONNECT"
	JACKIE_DISCONNECT = "DISCONNECT"
	JACKIE_MESSAGE    = "MESSAGE"
)

func ParseJackieRequest(b *bytes.Buffer) (string, []string, error) {
	message := strings.Split(b.String(), " ")

	if len(message) < 2 {
		return "", nil, errors.New("incorrect jackie length protocol")
	}

	proto := message[0]
	act := message[1]

	if proto == "JACKIE" {
		return act, message[2:], nil
	}

	return "", nil, errors.New("invalid jackie protocol")
}
