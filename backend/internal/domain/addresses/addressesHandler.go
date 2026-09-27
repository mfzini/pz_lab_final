package addresses

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type handler struct {
	*service
}

func NewHandler(service *service) *handler {
	return &handler{service}
}

func (this *handler) Setup(mux *http.ServeMux) {

	mux.HandleFunc("POST /direcciones", this.create)

	mux.HandleFunc("GET /direcciones/{id}", this.getByID)
	mux.HandleFunc("GET /direcciones", this.search)

}

func (this *handler) search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	personID := query.Get("persona")
	if personID != "" {
		this.getByPersonID(w, r, personID)
		return
	}

	street := query.Get("calle")
	house := query.Get("casa")
	if street != "" && house != "" {
		this.getByStreetAndHouse(w, r, street, house)
		return
	}

	w.WriteHeader(400)
}

func (this *handler) getByPersonID(w http.ResponseWriter, r *http.Request, id string) {
	addresses, err := this.service.repo.FindByPersonID(r.Context(), id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(500)
		return
	}
	if addresses == nil {
		w.WriteHeader(404)
		return
	}

	json.NewEncoder(w).Encode(addresses)
}

func (this *handler) getByStreetAndHouse(w http.ResponseWriter, r *http.Request, street, house string) {
	address, err := this.service.repo.FindByStreetAndHouse(r.Context(), street, house)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(500)
		return
	}
	if address == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(address)
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	addresses, err := this.service.repo.FindById(r.Context(), id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(500)
		return
	}

	if addresses == nil {
		w.WriteHeader(404)
		return
	}

	json.NewEncoder(w).Encode(addresses)
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	address := &Model{}

	if err := json.NewDecoder(r.Body).Decode(address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(400)
		return
	}

	err := this.service.Create(address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(address)
}
