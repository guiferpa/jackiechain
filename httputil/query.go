package httputil

import (
	"net/http"
	"strconv"
)

func GetQueryInt(req *http.Request, name string, defa int) int {
	if !req.URL.Query().Has(name) {
		return defa
	}

	v, err := strconv.Atoi(req.URL.Query().Get(name))
	if err != nil {
		return defa
	}
	return v
}
