package users

import "example/internal/domain/common"

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
