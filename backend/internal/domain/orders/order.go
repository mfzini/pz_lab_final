package orders

import (
	"example/internal/domain/addresses"
	"example/internal/domain/common"
	"example/internal/domain/products"
	"example/internal/domain/users"
)

// Estados de una orden, en orden de avance. Para agregar uno nuevo alcanza con
// sumarlo acá y al check de la tabla orders (script_pg.sql).
const (
	// StatusPending recién creada, todavía nadie la tocó.
	StatusPending = "pending"
	// StatusPreparing en proceso de preparación.
	StatusPreparing = "preparing"
	// StatusShipped ya se despachó.
	StatusShipped = "shipped"
)

// ValidStatus dice si el estado es uno conocido.
func ValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusPreparing, StatusShipped:
		return true
	}
	return false
}

type Model struct {
	Id,
	Status string
	// La columna es anulable: sin detalle se lee null, no "".
	Detail   *string
	Client   *users.Model
	Products []*products.Model
	Address  *addresses.Model
	Total    int
	common.TimeStamps
}
