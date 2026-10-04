package addresses

import "example/internal/domain/common"

type Model struct {
	Id,
	Street,
	House,
	Floor *int
	Coordinates,
	Detail,
	Door *string
	common.TimeStamps
}
