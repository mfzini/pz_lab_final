package controllers

import (
	"encoding/json"
	"example/internal/models"
	"example/internal/services"
	"fmt"
	"net/http"
)

type Response struct {
	Location string
}

func HandlePeople(mux *http.ServeMux, service *services.PeopleService) {
	mux.HandleFunc("GET /person", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		email := query.Get("email")
		p, err := service.Repo.FindByEmail(email)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		if p == nil {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(p)
	})

	mux.HandleFunc("GET /person/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p, err := service.Repo.FindById(id)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		if p == nil {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(p)
	})

	mux.HandleFunc("POST /person", func(w http.ResponseWriter, r *http.Request) {
		p := &models.Person{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		err = service.Create(p)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(&Response{
			Location: fmt.Sprintf("/person/%s", p.Id)})
	})
}
