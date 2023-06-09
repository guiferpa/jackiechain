package tcp

import (
	"bufio"
	"net"
	"net/http"
)

func ParseHTTPRequest(conn net.Conn) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReader(conn))
}
