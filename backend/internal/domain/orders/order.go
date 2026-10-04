package orders

import (
	"example/internal/domain/addresses"
	"example/internal/domain/common"
	"example/internal/domain/people"
	"example/internal/domain/products"
)

type Model struct {
	Id,
	Detail,
	Status string
	Client   *people.Model
	Products []*products.Model
	Address  *addresses.Model
	Total    float64
	common.TimeStamps
}
