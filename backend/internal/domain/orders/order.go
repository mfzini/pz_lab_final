package orders

import (
	"example/internal/domain/common"
	"example/internal/domain/dishes"
	"example/internal/domain/people"
)

type Model struct {
	Id,
	Detail,
	Status string
	Client *people.Model
	Dishes []*dishes.Model
	common.TimeStamps
}
