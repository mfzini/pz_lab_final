package orders

import (
	"encoding/json"
	"net/http"
	"strings"

	"example/internal/api"
	"example/internal/domain/common"
	"example/internal/domain/users"
	"example/internal/security"
)

type handler struct {
	*service
}

func NewHandler(service *service) *handler {
	return &handler{service}
}

// StatusRequest acepta {"Status": "..."} y {"status": "..."}: Go decodifica
// los campos sin distinguir mayúsculas.
type StatusRequest struct {
	Status string
}

type StatusResponse struct {
	Id     string
	Status string
}

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("POST /ordenes", security.AllowClients(this.create))
	// Historial del usuario logueado.
	mux.HandleFunc("GET /ordenes", security.AllowClients(this.listMine))
	// Vista del admin: todas las órdenes, filtrables por fecha y estado.
	mux.HandleFunc("GET /ordenes/admin", security.OnlyAdmin(this.listAll))
	mux.HandleFunc("GET /ordenes/{id}", security.AllowClients(this.getByID))
	mux.HandleFunc("PATCH /ordenes/{id}/estado", security.OnlyAdmin(this.updateStatus))
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	claims := security.GetClaimsFromContext(r.Context())

	order := &Model{}
	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// El cliente sale del token: nunca del body.
	order.Client = &users.Model{Id: claims.Subject}

	if err := this.service.Create(r.Context(), order); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, order, http.StatusCreated)
}

// listMine devuelve el historial paginado de las órdenes del usuario logueado.
// GET /ordenes?page=1&size=10
func (this *handler) listMine(w http.ResponseWriter, r *http.Request) {
	claims := security.GetClaimsFromContext(r.Context())
	page, size := common.PageParams(r.URL.Query())

	orders, pages, err := this.service.FindByUser(r.Context(), claims.Subject, page, size)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        orders,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

// listAll es la vista del admin: todas las órdenes con filtros y paginación.
// GET /ordenes/admin?status=pending&from=2026-10-01&to=2026-10-20&page=1&size=10
func (this *handler) listAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, size := common.PageParams(query)

	orders, pages, err := this.service.FindAll(r.Context(),
		query.Get("status"), query.Get("from"), query.Get("to"), page, size)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        orders,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	claims := security.GetClaimsFromContext(r.Context())

	order, err := this.service.FindById(r.Context(), r.PathValue("id"))
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	if order == nil {
		api.WriteJson(w, api.NewError("Orden no encontrada.", http.StatusNotFound), http.StatusNotFound)
		return
	}
	// El dueño ve la suya; el admin ve todas.
	if order.Client.Id != claims.Subject && claims.Role != security.Admin {
		api.WriteJson(w, api.NewError("No podés ver esta orden.", http.StatusForbidden), http.StatusForbidden)
		return
	}
	api.WriteJson(w, order, http.StatusOK)
}

// updateStatus es el flujo de despacho: sólo el admin avanza el estado.
func (this *handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	req := &StatusRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if err := this.service.UpdateStatus(r.Context(), id, status); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &StatusResponse{Id: id, Status: status}, http.StatusOK)
}
