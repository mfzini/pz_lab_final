package addresses

import "example/internal/domain/common"

type Model struct {
	Id,
	User_ID,
	Street,
	House string // la casa puede ser "12" o "12B": siempre texto
	Floor *int
	Coordinates,
	Detail,
	Door *string
	common.TimeStamps
}
