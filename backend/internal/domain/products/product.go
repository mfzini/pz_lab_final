package products

import (
	"example/internal/domain/common"
)

type Model struct {
	Id,
	Name string
	Price  int // centavos: 1250 = $12,50 (al mostrar se divide por 100)
	Amount int
	// La columna es anulable: sin foto se lee null, no "".
	Portrait *string
	// Categoría (anulable) y su nombre, que viene de un join en las lecturas.
	CategoryId   *string
	CategoryName *string
	common.TimeStamps
}
