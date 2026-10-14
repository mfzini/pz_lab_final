package users

import (
	"context"
	"example/internal/apiErrors"
	"fmt"
	"net/http"
	"os"
	"uuid"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *repo
}

func NewService(repo *repo) *Service {
	return &Service{repo}
}

func (this *Service) Create(user *Model) *apiErrors.Error {
	if user.Id == "" {
		user.Id = uuid.NewV7().String()
	}
	if user.Name == "" {
		return apiErrors.New("El nombre no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Surname == "" {
		return apiErrors.New("El apellido no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Email == "" {
		return apiErrors.New("El email no puede estar vacio.", http.StatusBadRequest)
	}

	if exists, _ := this.repo.FindByEmail(user.Email); exists != nil {
		return apiErrors.New("Email no válido.", http.StatusNotAcceptable)
	}
	if err := this.repo.create(user); err != nil {
		return apiErrors.Generic(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Muy grande me lastima. %v\n", err)
		return apiErrors.FromError(err, http.StatusBadRequest)
	}
	user.Password = string(hash)
	return nil
}

func (this *Service) Update(user *Model) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Muy grande me lastima. %v\n", err)
		return apiErrors.FromError(err, http.StatusBadRequest)
	}
	user.Password = string(hash)
	return this.repo.update(user)
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
