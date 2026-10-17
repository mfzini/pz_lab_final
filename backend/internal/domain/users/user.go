package users

import (
	"example/internal/domain/common"
	"example/internal/security"

	"golang.org/x/crypto/bcrypt"
)

type Model struct {
	Id,
	Name,
	Surname,
	Email,
	Phone string
	Role     security.Role
	Avatar   *string
	Password string `json:"-"`
	common.TimeStamps
}

func (this *Model) hashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(this.Password), 10)
	if err != nil {
		return err
	}
	this.Password = string(hash)
	return nil
}

func (this *Model) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(this.Password), []byte(password))
}
