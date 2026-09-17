package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hola, mundo!"))
	})
	http.ListenAndServe(":3000", nil)
}
