package controllers

import (
	"encoding/json"
	"example/internal/models"
	"example/internal/services"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type DishesController struct {
	mux     *http.ServeMux
	Service *services.DishesService
}

func NewDishesController(mux *http.ServeMux, service *services.DishesService) *DishesController {

	controller := &DishesController{
		mux:     mux,
		Service: service,
	}

	mux.HandleFunc("POST /platos/crear", func(w http.ResponseWriter, r *http.Request) {
		// esto deberiamos moverlo a un middleware independiente
		file, header, err := r.FormFile("picture")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dstPath := filepath.Join("uploads", header.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		file.Close()
		///////////////////////////////////////////////////////////

		dish := &models.Dish{}
		json.Unmarshal([]byte(r.FormValue("dish")), dish)
		service.Create(r.Context(), dish)

	})

	return controller
}
