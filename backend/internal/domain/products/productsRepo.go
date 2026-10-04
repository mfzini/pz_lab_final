package products

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type PRODUCT_SORTER string

const (
	SORT_BY_NAME     PRODUCT_SORTER = "name"
	SORT_BY_PRICE    PRODUCT_SORTER = "price"
	SORT_BY_CREATION PRODUCT_SORTER = "created_at"
)

type repo struct {
	db *pgx.Conn
}

func NewRepo(connection *pgx.Conn) *repo {
	return &repo{db: connection}
}

func (this *repo) create(ctx context.Context, product *Model) error {
	query := `insert into products (id, name, price, portrait) values (@product_id, @product_name, @price, @portrait);`

	_, err := this.db.Exec(ctx, query, pgx.NamedArgs{
		"product_id":   product.Id,
		"product_name": product.Name,
		"price":        product.Price,
		"portrait":     product.Portrait,
		"amount":       product.Amount,
	})
	return err
}

func (this *repo) update(ctx context.Context, product *Model) error {
	query := `update products set name = $1, price = $2, amount = $3, updated_at=$4`
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query,
		product.Name,
		product.Price,
		product.Amount,
		timestamp)
	if err == nil {
		product.UpdatedAt = &timestamp
	}
	return err
}

func (this *repo) delete(ctx context.Context, product *Model) error {
	query := "update products set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query, product.Id, timestamp)
	if err == nil {
		product.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) findByName(ctx context.Context, name string) ([]*Model, error) {
	query := `select * from products where name ilike $1`
	r, err := this.db.Query(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) findById(ctx context.Context, id string) (*Model, error) {
	query := `select * from products where id = $1`
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectExactlyOneRow(r, pgx.RowToAddrOfStructByNameLax[Model])
}

func (this *repo) list(ctx context.Context, page, size int, sorter PRODUCT_SORTER) ([]*Model, int, error) {
	offset := (page - 1) * size
	fromWhere := "from products where deleted_at is null"

	query := fmt.Sprintf("select * %s order by %s limit $1 offset $2;", fromWhere, sorter)
	rows, err := this.db.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}

	products, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Model])
	if err != nil {
		return nil, 0, err
	}

	countQuery := "select count(*) " + fromWhere
	totalRows := 0
	totalPages := 0
	err = this.db.QueryRow(context.Background(), countQuery).Scan(&totalRows)
	if size > 0 {
		totalPages = (totalRows + size - 1) / size
	}

	return products, totalPages, nil
}

func (this *repo) listByName(ctx context.Context, name string, page, size int, sorter PRODUCT_SORTER) ([]*Model, int, error) {
	offset := (page - 1) * size
	fromWhere := "from products where deleted_at is null and name ilike $1"
	query := fmt.Sprintf("select * %s order by %s limit $2 offset $3;", fromWhere, sorter)

	rows, err := this.db.Query(ctx, query, "%"+name+"%", size, offset)
	if err != nil {
		fmt.Println("meow 1")
		return nil, 0, err
	}
	products, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		fmt.Println("meow 2")
		return nil, 0, err
	}

	countQuery := "select count(*) " + fromWhere
	totalRows := 0
	totalPages := 0
	err = this.db.QueryRow(ctx, countQuery, "%"+name+"%").Scan(&totalRows)
	if err != nil {
		return nil, 0, err
	}

	totalPages = (totalRows + size - 1) / size
	return products, totalPages, nil
}
