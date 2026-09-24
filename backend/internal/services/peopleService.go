package services

import (
	"example/internal/models"
	"example/internal/repositories"
	"uuid"
)

type PeopleService struct {
	Repo *repositories.PeopleRepo
}

func (self *PeopleService) Create(p *models.Person) error {
	if p.Id == "" {
		p.Id = uuid.NewV7().String()
	}
	return self.Repo.Create(p)
}
