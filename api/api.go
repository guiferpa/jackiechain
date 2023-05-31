package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/guiferpa/jackchain/blockchain"
	"github.com/unrolled/render"
)

func Run(chain blockchain.Chain, port string) {
	router := chi.NewRouter()

	rnd := render.New()

	router.Get("/chain", func(w http.ResponseWriter, r *http.Request) {
		rnd.JSON(w, http.StatusOK, chain)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}
