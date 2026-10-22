package categories

import (
	"context"
	"fmt"

	"example/internal/domain/common"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Columnas de lectura, siempre iguales (prefijo implícito: no hay join).
const columns = `id, name, created_at, updated_at`

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repo {
	return &repo{db}
}

// create deja el id y el created_at que genera la DB en el modelo: el front
// necesita el id recién creado para asignar la categoría al plato.
func (this *repo) create(ctx context.Context, category *Model) error {
	query := `insert into categories (name) values ($1) returning id, created_at`
	return this.db.QueryRow(ctx, query, category.Name).Scan(&category.Id, &category.CreatedAt)
}

func (this *repo) list(ctx context.Context, name string, page, size int) ([]*Model, int, error) {
	where := "where deleted_at is null"
	args := []any{}
	if name != "" {
		args = append(args, "%"+name+"%")
		where += fmt.Sprintf(" and name ilike $%d", len(args))
	}

	args = append(args, size, (page-1)*size)
	query := fmt.Sprintf(`select %s from categories %s order by name
		limit $%d offset $%d`, columns, where, len(args)-1, len(args))
	rows, err := this.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	categories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		return nil, 0, err
	}

	countArgs := args[:len(args)-2]
	countQuery := fmt.Sprintf(`select count(*) from categories %s`, where)
	totalRows := 0
	if err := this.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return categories, common.TotalPagesFor(totalRows, size), nil
}
