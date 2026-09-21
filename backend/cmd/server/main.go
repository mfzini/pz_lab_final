package main

import (
	"example/internal/controllers"
	"example/internal/middleware"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	middleware.UseJson(mux)
	middleware.UseCors(mux)

	controllers.SetupPersonHandlers(mux)

	http.ListenAndServe(":3000", mux)
}
