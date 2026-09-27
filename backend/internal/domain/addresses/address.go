package addresses

import "example/internal/domain/common"

type Model struct {
	Id,
	Street,
	House,
	Coordinates *string
	Floor *int
	Door  *string
	common.TimeStamps
}
