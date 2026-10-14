package orders

import (
	"example/internal/domain/addresses"
	"example/internal/domain/common"
	"example/internal/domain/products"
	"example/internal/domain/users"
)

type Model struct {
	Id,
	Detail,
	Status string
	Client   *users.Model
	Products []*products.Model
	Address  *addresses.Model
	Total    int
	common.TimeStamps
}
