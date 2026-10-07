package products

import (
	"example/internal/domain/common"
)

type Model struct {
	Id,
	Name,
	Portrait string
	Price  int
	Amount int
	common.TimeStamps
}
