package addresses

import "example/internal/domain/common"

type Model struct {
	ID,
	User_ID,
	Street string
	House,
	Floor *int
	Coordinates,
	Detail,
	Door *string
	common.TimeStamps
}
