package products

import (
	"example/internal/domain/common"
)

type Model struct {
	Id,
	Name,
	Portrait string
	Price  float64
	Amount int
	common.TimeStamps
}
