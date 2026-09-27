package dishes

import (
	"encoding/json"
	"example/internal/domain/images"
	"fmt"
	"net/http"
	"os"
)

type handler struct {
	service *service
}

func NewHandler(service *service) *handler {
	return &handler{service}
}

func (this *handler) Setup(mux *http.ServeMux) {

	mux.HandleFunc("POST /platos", this.postDish)
}

func (this *handler) postDish(w http.ResponseWriter, r *http.Request) {
	dish := &Model{}
	if err := json.Unmarshal([]byte(r.FormValue("dish")), dish); err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("picture")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	dish.Portrait = &images.Model{
		Multipart:    &file,
		OriginalName: header.Filename,
	}
	err = this.service.Create(r.Context(), dish)

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(dish)
}
