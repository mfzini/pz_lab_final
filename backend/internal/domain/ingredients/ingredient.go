package ingredients

import "example/internal/domain/common"

type Ingredient struct {
	Id,
	Name string
	Stock,
	Ammount float32
	common.TimeStamps
}
