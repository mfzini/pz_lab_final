package common

import "time"

// CreatedAt/UpdatedAt van en el JSON: el front los muestra en el historial y
// en el listado del admin (que filtra justamente por fecha). DeletedAt no.
type TimeStamps struct {
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}
