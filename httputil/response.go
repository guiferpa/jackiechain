package httputil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func newHTTPResponse(req *http.Request, statusCode int, header http.Header, body *bytes.Buffer) *http.Response {
	var rdb io.ReadCloser = http.NoBody

	if body.Len() > 0 {
		rdb = io.NopCloser(body)
	}

	resp := &http.Response{
		StatusCode:       statusCode,
		ProtoMajor:       1,
		ProtoMinor:       1,
		Request:          req,
		TransferEncoding: []string{"utf8"},
		Trailer:          header,
		Body:             rdb,
		Header:           header,
		ContentLength:    int64(body.Len()),
	}

	return resp
}

func Response(r *http.Request, w io.Writer, statusCode int, body *bytes.Buffer) error {
	header := make(http.Header, 0)
	header.Set("Content-Type", "application/json; charset=utf8")
	header.Set("Date", time.Now().Format(time.RFC1123))

	resp := newHTTPResponse(r, statusCode, header, body)

	bs := &bytes.Buffer{}
	if err := resp.Write(bs); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, bs)
	if err == io.EOF {
		return nil
	}

	return err
}

func ResponseNotFound(r *http.Request, w io.Writer) error {
	body := "{\"message\": \"Not found\"}"
	return Response(r, w, http.StatusNotFound, bytes.NewBufferString(body))
}
