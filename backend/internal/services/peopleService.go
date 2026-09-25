package services

import (
	"example/internal/common"
	"example/internal/models"
	"example/internal/repositories"
	"fmt"
	"uuid"
)

type PeopleService struct {
	Repo *repositories.PeopleRepo
}

func (self *PeopleService) Create(p *models.Person) *common.ApiError {
	if p.Id == "" {
		p.Id = uuid.NewV7().String()
	}
	if err := validateFieldNotEmpty("name", p.Name); err != nil {
		return err
	}
	if err := validateFieldNotEmpty("surname", p.Surname); err != nil {
		return err
	}
	if err := validateFieldNotEmpty("email", p.Email); err != nil {
		return err
	}
	if exists, _ := self.Repo.FindByEmail(p.Email); exists != nil {
		return &common.ApiError{
			Err:        "Email no válido.",
			StatusCode: 400,
		}
	}
	if err := self.Repo.Create(p); err != nil {
		return &common.ApiError{
			Err:        err.Error(),
			StatusCode: 500,
		}
	}
	return nil
}

func (self *PeopleService) Update(p *models.Person) error {
	return nil
}
func validateFieldNotEmpty(field, input string) *common.ApiError {
	if input == "" {
		return &common.ApiError{
			Err:        fmt.Sprintf("El %s no puede estar vacio.", field),
			StatusCode: 400,
		}
	}
	return nil
}
