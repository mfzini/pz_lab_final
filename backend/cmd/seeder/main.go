package main

import (
	"example/internal/models"
	"example/internal/repositories"
)

func main() {
	peopleRepo := &repositories.PeopleRepo{}

	lionel := models.NewPerson(
		"Lionel",
		"Hutz",
		"lionel_hutz@spring.field",
		"123456")
	cosme := models.NewPerson(
		"Cosme",
		"Fulanito",
		"cosme_fulanito@spring.field",
		"123456")
	peopleRepo.Save(lionel)
	peopleRepo.Save(cosme)
}
