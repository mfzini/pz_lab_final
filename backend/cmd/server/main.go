package main

import (
	"example/internal/db"
	"example/internal/domain/addresses"
	"example/internal/domain/dishes"
	"example/internal/domain/images"
	"example/internal/domain/people"
	"example/internal/middleware"

	"net/http"
)

func main() {
	Init()
}

func Init() {
	mux := http.NewServeMux()

	imagesRepo := images.NewRepo(db.GetConnection())
	imagesService := images.NewService(imagesRepo)
	images.ServeWithMux(mux)

	peopleHandler := people.NewHandler(people.NewService(people.NewRepo(db.GetConnection())))
	peopleHandler.Setup(mux)

	addressesHandler := addresses.NewHandler(addresses.NewService(addresses.NewRepo(db.GetConnection())))
	addressesHandler.Setup(mux)

	dishesHandler := dishes.NewHandler(dishes.NewService(dishes.NewRepo(db.GetConnection(), imagesService)))
	dishesHandler.Setup(mux)

	http.ListenAndServe(":3000", middleware.Cors(mux))
}
