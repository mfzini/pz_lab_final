package products

import (
	"encoding/json"
	"example/internal/domain/common"
	"example/internal/security"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	mux.HandleFunc("PATCH /platos/{id}", security.OnlyAdmin(this.getproduct))
}

func (this *handler) updateProduct(w http.ResponseWriter, r *http.Request) {
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
	if err = this.service.Update(r.Context(), product, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(product)
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

	err = this.service.Create(r.Context(), product, file)

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(product)
}

func (this *handler) getproduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	product, err := this.service.FindById(r.Context(), id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(product)
}

func (this *handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		page = 1
	}
	size, err := strconv.Atoi(query.Get("size"))
	if err != nil {
		size = 10
	}
	name := query.Get("name")
	sortBy := query.Get("sortby")

	if list, pages, err := this.service.List(r.Context(), name, sortBy, page, size); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(&common.PageResponse[*Model]{
			Data:        list,
			CurrentPage: page,
			TotalPages:  pages})
	}
}
