package products

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Columnas de lectura, siempre iguales. Con el join a categories hay que
// prefijar todo: "name" existe en ambas tablas (ambiguo) y "p.*" traería
// columnas de sobra que el struct no tiene. deleted_at va incluido porque el
// scan estricto (RowToStructByName) exige todos los campos del struct.
const columns = `p.id, p.name, p.price, p.amount, p.portrait,
	p.category_id, c.name as category_name, p.created_at, p.updated_at, p.deleted_at`

const join = `left join categories c on c.id = p.category_id`

type PRODUCT_SORTER string

const (
	SORT_BY_NAME     PRODUCT_SORTER = "name"
	SORT_BY_PRICE    PRODUCT_SORTER = "price"
	SORT_BY_CREATION PRODUCT_SORTER = "created_at"
)

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(connection *pgxpool.Pool) *repo {
	return &repo{db: connection}
}

func (this *repo) create(ctx context.Context, product *Model) error {
	query := `insert into products (id, name, price, portrait, category_id)
		values (@product_id, @product_name, @price, @portrait, @category_id);`

	_, err := this.db.Exec(ctx, query, pgx.NamedArgs{
		"product_id":   product.Id,
		"product_name": product.Name,
		"price":        product.Price,
		"portrait":     product.Portrait,
		"category_id":  product.CategoryId,
	})
	return err
}

func (this *repo) update(ctx context.Context, product *Model) (bool, error) {
	// coalesce: sin foto nueva (nil) se mantiene la que el plato ya tenía.
	// category_id sí se asigna tal cual: mandar null en el body la limpia.
	query := `update products set name = $1, price = $2, portrait = coalesce($3, portrait),
		category_id = $4, updated_at = $5 where id = $6 and deleted_at is null`
	timestamp := time.Now()
	tag, err := this.db.Exec(ctx, query,
		product.Name,
		product.Price,
		product.Portrait,
		product.CategoryId,
		timestamp,
		product.Id,
	)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() > 0 {
		product.UpdatedAt = &timestamp
	}
	return tag.RowsAffected() > 0, nil
}

func (this *repo) delete(ctx context.Context, product *Model) error {
	query := "update products set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query, timestamp, product.Id)
	if err == nil {
		product.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) findById(ctx context.Context, id string) (*Model, error) {
	query := `select ` + columns + ` from products p ` + join + `
		where p.id = $1 and p.deleted_at is null`
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectExactlyOneRow(r, pgx.RowToAddrOfStructByNameLax[Model])
}

func (this *repo) list(ctx context.Context, page, size int, sorter PRODUCT_SORTER) ([]*Model, int, error) {
	offset := (page - 1) * size
	fromWhere := "from products p " + join + " where p.deleted_at is null"

	query := fmt.Sprintf("select %s %s order by p.%s limit $1 offset $2;", columns, fromWhere, sorter)
	rows, err := this.db.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}

	products, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Model])
	if err != nil {
		return nil, 0, err
	}

	countQuery := "select count(*) from products p where p.deleted_at is null"
	totalRows := 0
	totalPages := 0
	if err := this.db.QueryRow(ctx, countQuery).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	if size > 0 {
		totalPages = (totalRows + size - 1) / size
	}

	return products, totalPages, nil
}

func (this *repo) listByName(ctx context.Context, name string, page, size int, sorter PRODUCT_SORTER) ([]*Model, int, error) {
	offset := (page - 1) * size
	where := "where p.deleted_at is null and p.name ilike $1"
	query := fmt.Sprintf("select %s from products p %s %s order by p.%s limit $2 offset $3;",
		columns, join, where, sorter)

	rows, err := this.db.Query(ctx, query, "%"+name+"%", size, offset)
	if err != nil {
		return nil, 0, err
	}
	products, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		return nil, 0, err
	}

	countQuery := "select count(*) from products p " + where
	totalRows := 0
	totalPages := 0
	if err := this.db.QueryRow(ctx, countQuery, "%"+name+"%").Scan(&totalRows); err != nil {
		return nil, 0, err
	}

	totalPages = (totalRows + size - 1) / size
	return products, totalPages, nil
}
