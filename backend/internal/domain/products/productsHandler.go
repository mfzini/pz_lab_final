package products

import (
	"encoding/json"
	"errors"
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

type orderRequest struct {
	productes []*Model
}

func NewHandler(service *Service) *handler {
	return &handler{service}
}

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("POST /platos", security.OnlyAdmin(this.postProduct))
	mux.HandleFunc("GET /platos/list", this.list)
	mux.HandleFunc("GET /platos/{id}", this.getproduct)
	mux.HandleFunc("PATCH /platos/{id}", security.OnlyAdmin(this.updateProduct))
}

func (this *handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	product := &Model{}
	if err := json.Unmarshal([]byte(r.FormValue("product")), product); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// El id sale del path: lo que mande el body no decide a qué plato se le pega.
	product.Id = r.PathValue("id")

	// La foto es opcional al editar: sin archivo, el service mantiene la actual.
	file, _, err := r.FormFile("picture")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if file != nil {
		defer file.Close()
	}

	if err := this.service.Update(r.Context(), product, file); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, product, http.StatusOK)
}

func (this *handler) postProduct(w http.ResponseWriter, r *http.Request) {
	product := &Model{}
	if err := json.Unmarshal([]byte(r.FormValue("product")), product); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("picture")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := this.service.Create(r.Context(), product, file); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, product, http.StatusCreated)
}

func (this *handler) getproduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	product, err := this.service.FindById(r.Context(), id)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, product, http.StatusOK)
}

func (this *handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, size := common.PageParams(query)
	name := query.Get("name")
	sortBy := query.Get("sortby")

	list, pages, err := this.service.List(r.Context(), name, sortBy, page, size)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        list,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}
