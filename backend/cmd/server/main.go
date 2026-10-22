package main

import (
	"example/internal/db"
	"example/internal/domain/addresses"
	"example/internal/domain/categories"
	"example/internal/domain/images"
	"example/internal/domain/orders"
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

	jwtService := security.NewJWTService()
	usersRepo := users.NewRepo(db.GetConnection())
	usersService := users.NewService(usersRepo, imagesService)
	usersHandler := users.NewHandler(usersService, jwtService)
	usersHandler.Setup(mux)

	addressesRepo := addresses.NewRepo(db.GetConnection())
	addressesService := addresses.NewService(addressesRepo)
	addressesHandler := addresses.NewHandler(addressesService)
	addressesHandler.Setup(mux)

	productsRepo := products.NewRepo(db.GetConnection())
	productsService := products.NewService(productsRepo, imagesService)
	productsHandler := products.NewHandler(productsService)
	productsHandler.Setup(mux)

	// Maestra de categorías: el form de platos la busca por nombre (async).
	categoriesRepo := categories.NewRepo(db.GetConnection())
	categoriesService := categories.NewService(categoriesRepo)
	categoriesHandler := categories.NewHandler(categoriesService)
	categoriesHandler.Setup(mux)

	ordersRepo := orders.NewRepo(db.GetConnection())
	ordersService := orders.NewService(ordersRepo, productsService, addressesService)
	ordersHandler := orders.NewHandler(ordersService)
	ordersHandler.Setup(mux)

	mainHandler := jwtService.SetUserContext(mux)
	http.ListenAndServe(":3000", security.Cors(mainHandler))
}
