package tcp

import (
	"bufio"
	"bytes"
	"net/http"
)

func ParseHTTPRequest(b *bytes.Buffer) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReader(b))
}
