package main

import (
	"example/internal/controllers"
	"example/internal/db"
	"example/internal/middleware"
	"example/internal/repositories"
	"example/internal/services"
	"net/http"
)

func main() {
	Init()
}

func Init() {
	mux := http.NewServeMux()
	peopleService := &services.PeopleService{
		Repo: repositories.NewPeopleRepo(db.GetConnection()),
	}
	//dishesRepo := repositories.NewDishesRepo(db.GetConnection())

	controllers.HandlePeople(mux, peopleService)

	http.ListenAndServe(":3000", middleware.Cors(middleware.Json(mux)))
}
