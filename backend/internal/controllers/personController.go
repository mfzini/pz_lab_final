package controllers

import (
	"encoding/json"
	"example/internal/repositories"
	"net/http"
)

func SetupPersonHandlers(mux *http.ServeMux) {

	repo := &repositories.PeopleRepo{}

	mux.HandleFunc("GET /person", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		email := query.Get("email")
		p, err := repo.FindByEmail(email)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"Algo malió sal!"}`))
			return
		}
		if p == nil {
			w.WriteHeader(404)
			w.Write([]byte(`{"error":"Person not found!"}`))
			return
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(p)
	})

}

var repo = &repositories.PeopleRepo{}

var FindByEmail = func(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	email := query.Get("email")
	p, err := repo.FindByEmail(email)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Algo malió sal!"}`))
		return
	}
	if p == nil {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Person not found!"}`))
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(p)
}
