package dishes

import (
	"example/internal/domain/common"
	"example/internal/domain/images"
)

type Model struct {
	Id,
	Name string
	Price    float64
	Portrait *images.Model
	common.TimeStamps
}
