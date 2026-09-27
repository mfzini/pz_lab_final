package people

import (
	"example/internal/apiErrors"
	"net/http"
	"uuid"
)

type service struct {
	repo *repo
}

func NewService(repo *repo) *service {
	return &service{repo}
}

func (this *service) Create(p *Model) *apiErrors.Error {
	if p.Id == "" {
		p.Id = uuid.NewV7().String()
	}
	if p.Name == "" {
		return apiErrors.New("El nombre no puede estar vacio.", http.StatusBadRequest)
	}
	if p.Surname == "" {
		return apiErrors.New("El apellido no puede estar vacio.", http.StatusBadRequest)
	}
	if p.Email == "" {
		return apiErrors.New("El email no puede estar vacio.", http.StatusBadRequest)
	}

	if exists, _ := this.repo.FindByEmail(p.Email); exists != nil {
		return apiErrors.New("Email no válido.", http.StatusNotAcceptable)
	}
	if err := this.repo.create(p); err != nil {
		return apiErrors.Generic(err)
	}
	return nil
}

func (this *service) Update(p *Model) error {
	return nil
}
