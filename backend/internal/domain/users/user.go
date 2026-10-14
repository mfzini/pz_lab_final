package users

import (
	"example/internal/domain/common"

	"golang.org/x/crypto/bcrypt"
)

type Model struct {
	Id,
	Name,
	Surname,
	Email,
	Role string
	Phone,
	Avatar *string
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
