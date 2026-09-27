package dishes

import (
	"context"
	"example/internal/domain/images"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type DISH_SORTER string

const (
	SORT_BY_NAME     DISH_SORTER = "name"
	SORT_BY_PRICE    DISH_SORTER = "price"
	SORT_BY_CREATION DISH_SORTER = "createdat"
)

type repo struct {
	db            *pgx.Conn
	imagesService *images.Service
}

func NewRepo(connection *pgx.Conn, imagesService *images.Service) *repo {
	return &repo{db: connection, imagesService: imagesService}
}

func (this *repo) create(ctx context.Context, dish *Model) error {
	tx, err := this.db.Begin(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := this.imagesService.Create(ctx, tx, dish.Portrait); err != nil {
		return err
	}

	query := `insert into dishes (id, name, price, portrait, created_at) values (@dish_id, @dish_name, @price, @portrait, @timestamp);`

	_, err = tx.Exec(ctx, query, pgx.NamedArgs{
		"dish_id":   dish.Id,
		"dish_name": dish.Name,
		"price":     dish.Price,
		"timestamp": dish.CreatedAt,
		"portrait":  dish.Portrait.Id,
	})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (this *repo) update(ctx context.Context, dish *Model) error {
	query := `update dishes set name = $1, price = $2, updated_at=$3`
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query,
		dish.Name,
		dish.Price,
		timestamp)
	if err == nil {
		dish.UpdatedAt = &timestamp
	}
	return err
}

func (this *repo) delete(ctx context.Context, dish *Model) error {
	query := "update dishes set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query, dish.Id, timestamp)
	if err == nil {
		dish.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindByName(ctx context.Context, name string) ([]*Model, error) {
	query := `select * from dishes where name ilike $1`
	r, err := this.db.Query(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) FindById(ctx context.Context, id string) (*Model, error) {
	query := `select * from dishes id = $1`
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) List(ctx context.Context, page, size int, sorter DISH_SORTER) (int, []*Model, error) {
	offset := (page - 1) * size
	where := "deletedat is not null"
	query := fmt.Sprintf(`select * from dishes
		where %s
        order by %s
        limit $2 offset $3;`, where, sorter)
	r, err := this.db.Query(ctx, query, sorter, size, offset)
	if err != nil {
		return 0, nil, err
	}
	defer r.Close()
	dishes, err := pgx.CollectRows(r, pgx.RowToAddrOfStructByName[Model])
	if err != nil {
		return 0, nil, err
	}
	pages, err := this.getCount(where)
	if err != nil {
		return 0, nil, err
	}
	return pages, dishes, nil
}

func (this *repo) getCount(where string) (int, error) {
	query := fmt.Sprintf("select count(*) as pages from dishes where %s", where)
	pages := 0
	err := this.db.QueryRow(context.Background(), query).Scan(&pages)
	return pages, err
}
