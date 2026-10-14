package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Response struct {
	Location string
}

type handler struct {
	*Service
}

func NewHandler(Service *Service) *handler {
	return &handler{Service}
}

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("GET /user/{id}", this.getByID)
	mux.HandleFunc("POST /user", this.create)
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	p := &Model{}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := this.Service.Create(p); err != nil {
		w.WriteHeader(err.StatusCode)
		fmt.Fprintln(os.Stderr, err)
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(&Response{
		Location: fmt.Sprintf("/person/%s", p.Id)})
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := this.Service.repo.FindById(id)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(os.Stderr, "%+v", err)
		return
	}
	if p == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(p)
}
