package main

import (
	"example/internal/db"
	"example/internal/domain/addresses"
	"example/internal/domain/images"
	"example/internal/domain/people"
	"example/internal/domain/products"
	"example/internal/middleware"
	"os"

	"net/http"
)

func main() {
	Init()
}

func Init() {
	mux := http.NewServeMux()

	imagesRepo := images.NewRepo(os.Getenv("PATH_UPLOADS"))
	imagesService := images.NewService(imagesRepo)
	images.ServeWithMux(mux)

	peopleHandler := people.NewHandler(people.NewService(people.NewRepo(db.GetConnection())))
	peopleHandler.Setup(mux)

	addressesHandler := addresses.NewHandler(addresses.NewService(addresses.NewRepo(db.GetConnection())))
	addressesHandler.Setup(mux)

	productsHandler := products.NewHandler(products.NewService(products.NewRepo(db.GetConnection()), imagesService))
	productsHandler.Setup(mux)

	http.ListenAndServe(":3000", middleware.Cors(mux))
}
