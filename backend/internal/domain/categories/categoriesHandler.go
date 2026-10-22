package categories

import (
	"encoding/json"
	"net/http"

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

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("GET /categorias", this.list)
	mux.HandleFunc("POST /categorias", security.OnlyAdmin(this.create))
}

// GET /categorias?name=...&page=&size= — listado paginado; name filtra (ilike)
// y es lo que usa la búsqueda asíncrona del form de platos.
func (this *handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, size := common.PageParams(query)

	categories, pages, err := this.service.List(r.Context(), query.Get("name"), page, size)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        categories,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

// POST /categorias — body: {"name": "Pizza"} (sólo admin)
func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	category := &Model{}
	if err := json.NewDecoder(r.Body).Decode(category); err != nil {
		api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := this.service.Create(r.Context(), category); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, category, http.StatusCreated)
}
