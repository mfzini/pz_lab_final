package main

import (
	"example/internal/controllers"
	"example/internal/middleware"
	"example/internal/server"
	"fmt"
	"net/http"
)

func main() {
	app := server.New()
	app.Use(middleware.CorsMiddleware)
	app.Use(middleware.JsonMiddleware)
	app.Use(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("meow")
	})
	app.Handle("GET /person", controllers.FindByEmail)

	app.ListenAndServe(":3000")
}
