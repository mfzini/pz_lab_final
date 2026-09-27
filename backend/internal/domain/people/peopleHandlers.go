package people

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
	*service
}

func NewHandler(service *service) *handler {
	return &handler{service}
}

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("GET /persona/{id}", this.getByID)
	mux.HandleFunc("POST /persona", this.create)
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	p := &Model{}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := this.service.Create(p); err != nil {
		w.WriteHeader(err.StatusCode)
		fmt.Fprintln(os.Stderr, err)
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(&Response{
		Location: fmt.Sprintf("/person/%s", p.Id)})
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	id := r.PathValue("id")

	p, err := this.service.repo.FindById(id)
	if err != nil {
		w.WriteHeader(500)
		encoder.Encode(err)
		return
	}
	if p == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(p)
}
