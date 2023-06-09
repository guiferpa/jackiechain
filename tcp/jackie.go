package tcp

import (
	"errors"
	"net"
	"strings"
)

const (
	JACKIE_CONNECT    = "CONNECT"
	JACKIE_DISCONNECT = "DISCONNECT"
	JACKIE_MESSAGE    = "MESSAGE"
)

func ParseJackieRequest(conn net.Conn) (string, []string, error) {
	buf := make([]byte, 512)
	_, err := conn.Read(buf)
	if err != nil {
		return "", nil, err
	}

	message := strings.Split(string(buf), " ")

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
