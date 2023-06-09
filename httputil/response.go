package httputil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func NewHTTPResponse(req *http.Request, statusCode int, body []byte) *http.Response {
	var rdb io.ReadCloser = http.NoBody

	if len(body) > 0 {
		rdb = io.NopCloser(bytes.NewReader(body))
	}

	resp := &http.Response{
		StatusCode:       http.StatusOK,
		ProtoMajor:       1,
		ProtoMinor:       1,
		Request:          req,
		TransferEncoding: []string{"utf8"},
		Trailer:          nil,
		Body:             rdb,
	}

	return resp
}

func Response(w io.Writer, resp *http.Response) (int, error) {
	bs := &bytes.Buffer{}

	if err := resp.Write(bs); err != nil {
		return 0, err
	}

	l, err := fmt.Fprint(w, bs)
	if err == io.EOF {
		return l, nil
	}

	return l, err
}
