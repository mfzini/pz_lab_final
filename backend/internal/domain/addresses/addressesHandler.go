package addresses

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"example/internal/api"
	"example/internal/domain/common"
	"example/internal/security"
)

type handler struct {
	service *Service
}

func NewHandler(s *Service) *handler {
	return &handler{service: s}
}

func (this *handler) Setup(server *http.ServeMux) {
	server.HandleFunc("POST /direcciones", security.AllowClients(this.create))
	// Todas las lecturas validan dueño o admin (ver guardOwner / service).
	server.HandleFunc("GET /direcciones/{id}", security.AllowClients(this.getByID))
	server.HandleFunc("GET /usuarios/{id}/direcciones", security.AllowClients(this.getByUserID))
	// Cruza direcciones de varios usuarios: sólo admin.
	server.HandleFunc("GET /buscar/direcciones", security.OnlyAdmin(this.getByStreetAndHouse))
	server.HandleFunc("PATCH /direcciones/{id}", security.AllowClients(this.update))
	server.HandleFunc("DELETE /direcciones/{id}", security.AllowClients(this.delete))
}

// guardOwner: dueño del recurso o admin. Si no, escribe el 403 y devuelve false.
func guardOwner(w http.ResponseWriter, r *http.Request, id string) bool {
	claims := security.GetClaimsFromContext(r.Context())
	if claims != nil && (claims.Subject == id || claims.Role == security.Admin) {
		return true
	}
	api.WriteJson(w, api.NewError("No tenés permisos para esta operación.", http.StatusForbidden), http.StatusForbidden)
	return false
}

// GET /buscar/direcciones?street=...&house=...&page=1&size=10 (sólo admin)
func (this *handler) getByStreetAndHouse(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, size := common.PageParams(query)

	addresses, pages, err := this.service.FindByStreetAndHouse(r.Context(),
		query.Get("street"), query.Get("house"), page, size)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        addresses,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

// GET /usuarios/{id}/direcciones?street=...&page=1&size=10 — dueño o admin.
// street filtra por calle: es la búsqueda asíncrona del checkout.
func (this *handler) getByUserID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// El usuario sólo lista sus propias direcciones; el admin, las que quiera.
	if !guardOwner(w, r, id) {
		return
	}
	query := r.URL.Query()
	page, size := common.PageParams(query)

	addresses, pages, err := this.service.FindByUserID(r.Context(), id, query.Get("street"), page, size)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        addresses,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	address, err := this.service.FindById(r.Context(), id)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	if address == nil {
		api.WriteJson(w, api.NewError("Dirección no encontrada.", http.StatusNotFound), http.StatusNotFound)
		return
	}
	api.WriteJson(w, address, http.StatusOK)
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	address := Model{}

	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := this.service.Create(r.Context(), &address); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, address, http.StatusCreated)
}

func (this *handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	data := Model{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := this.service.Update(r.Context(), id, &data); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, data, http.StatusOK)
}

func (this *handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := this.service.Delete(r.Context(), id); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (this *handler) search(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	address, err := this.service.FindById(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJson(w, address, http.StatusOK)
}
