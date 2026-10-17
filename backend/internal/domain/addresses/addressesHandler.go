package addresses

import (
	"encoding/json"
	"example/internal/api"
	"example/internal/security"
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

	mux.HandleFunc("POST /direcciones", security.AllowClients(this.create))

	mux.HandleFunc("GET /direcciones/{id}", security.AllowClients(this.getByID))
	mux.HandleFunc("PATCH /direcciones/{id}", security.AllowClients(this.update))
	mux.HandleFunc("DELETE /direcciones/{id}", security.AllowClients(this.delete))
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	address := &Model{}

	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := this.service.Create(r.Context(), address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJson(w, address, http.StatusCreated)
}

func (this *handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data := &Model{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := this.service.Update(r.Context(), id, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	api.WriteJson(w, data, http.StatusAccepted)
}

func (this *handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := this.service.Delete(r.Context(), id); err != nil {
		w.WriteHeader(err.StatusCode)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	address, err := this.service.FindById(r.Context(), id)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		return
	}
	api.WriteJson(w, address, http.StatusOK)
}

func (this *handler) search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userID := query.Get("user")
	if userID != "" {
		this.getByUserID(w, r, userID)
		return
	}

	street := query.Get("calle")
	house := query.Get("casa")
	if street != "" && house != "" {
		this.getByStreetAndHouse(w, r, street, house)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (this *handler) getByUserID(w http.ResponseWriter, r *http.Request, id string) {
	addresses, err := this.service.repo.FindByUserID(r.Context(), id)
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
