package api

import "net/http"

func Run() {
	if err := http.ListenAndServe(":3000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})); err != nil {
		panic(err)
	}
}
