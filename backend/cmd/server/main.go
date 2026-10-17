package main

import (
	"example/internal/db"
	"example/internal/domain/addresses"
	"example/internal/domain/images"
	"example/internal/domain/products"
	"example/internal/domain/users"
	"example/internal/security"
	"os"

	"net/http"
)

func main() {
	Setup()
}

func Setup() {
	mux := http.NewServeMux()
	imagesRepo := images.NewRepo(os.Getenv("PATH_UPLOADS"))
	imagesService := images.NewService(imagesRepo)
	images.ServeWithMux(mux)

	usersRepo := users.NewRepo(db.GetConnection())
	usersService := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersService)
	usersHandler.Setup(mux)

	addressesRepo := addresses.NewRepo(db.GetConnection())
	addressesService := addresses.NewService(addressesRepo)
	addressesHandler := addresses.NewHandler(addressesService)
	addressesHandler.Setup(mux)

	productsRepo := products.NewRepo(db.GetConnection())
	productsService := products.NewService(productsRepo, imagesService)
	productsHandler := products.NewHandler(productsService)
	productsHandler.Setup(mux)

	jwtService := security.NewJWTService()

	mainHandler := jwtService.SetUserContext(mux)
	http.ListenAndServe(":3000", security.Cors(mainHandler))
}
