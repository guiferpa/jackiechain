package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/unrolled/render"
)

const HTTP_INTERFACE_PROTOCOL = "http"

type HTTPInterface struct {
	Port   string
	Router chi.Router
}

func (i *HTTPInterface) Run(msgc, cmdc chan string, errc chan error) {
	rnd := render.New()

	i.Router.Get("/chain", func(w http.ResponseWriter, r *http.Request) {
		cmdc <- fmt.Sprintf("HTTP GET CHAIN")

		msg := <-msgc

		bs, err := base64.StdEncoding.DecodeString(msg)
		if err != nil {
			panic(err)
		}

		chain := make(map[string]interface{}, 0)
		if err = json.Unmarshal(bs, &chain); err != nil {
			panic(err)
		}

		rnd.JSON(w, http.StatusOK, chain)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", i.Port), i.Router); err != nil {
		errc <- err
	}
}

func NewHTTPInterace(port string) *HTTPInterface {
	router := chi.NewRouter()

	return &HTTPInterface{Port: port, Router: router}
}
