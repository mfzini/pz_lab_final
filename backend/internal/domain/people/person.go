package people

import "example/internal/domain/common"

type Model struct {
	Id,
	Name,
	Surname,
	Phone,
	Email string
	common.TimeStamps
}
