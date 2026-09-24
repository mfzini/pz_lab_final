package main

import (
	"context"
	"example/internal/controllers"
	"example/internal/db"
	"example/internal/repositories"
	"example/internal/services"
	"fmt"
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
	dishesRepo := repositories.NewDishesRepo(db.GetConnection())
	_, r, err := dishesRepo.List(context.Background(), 1, 10, "")
	if err != nil {
		fmt.Println(err)
	}
	for _, d := range r {
		fmt.Println(d)
	}
	controllers.HandlePeople(mux, peopleService)

	//http.ListenAndServe(":3000", middleware.Cors(middleware.Json(mux)))
}
