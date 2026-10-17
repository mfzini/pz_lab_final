package users

import (
	"context"
	"example/internal/api"
	"net/http"
	"net/mail"
	"uuid"
)

type Service struct {
	repo *repo
}

func NewService(repo *repo) *Service {
	return &Service{repo}
}

func (this *Service) Create(user *Model) *api.Error {
	if user.Id == "" {
		user.Id = uuid.NewV7().String()
	}
	if user.Name == "" {
		return api.NewError("El nombre no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Surname == "" {
		return api.NewError("El apellido no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Email == "" {
		return api.NewError("El email no puede estar vacio.", http.StatusBadRequest)
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		return api.NewError("Formato de email no válido.", http.StatusBadRequest)
	}

	if exists, _ := this.repo.FindByEmail(user.Email); exists != nil {
		return api.NewError("Email no válido.", http.StatusNotAcceptable)
	}

	if err := user.hashPassword(); err != nil {
		return api.NewError("No se pudo procesar la contraseña.", http.StatusInternalServerError)
	}
	if err := this.repo.create(user); err != nil {
		return api.NewGenericError(err)
	}

	return nil
}

func (this *Service) Update(ctx context.Context, id string, user *Model) *api.Error {
	if err := this.repo.update(user); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *Service) FindById(id string) (*Model, error) {
	return this.repo.FindById(id)
}
func (this *Service) FindByEmail(email string) (*Model, error) {
	return this.repo.FindByEmail(email)
}

func (this *Service) AdminExists() bool {
	count := 0
	r := this.repo.db.QueryRow(context.Background(), "select count(*) from users where role = 'admin';")
	r.Scan(&count)
	return count > 0
}
