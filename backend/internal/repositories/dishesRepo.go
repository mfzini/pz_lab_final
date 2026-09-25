package repositories

import (
	"context"
	"example/internal/models"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type DISH_SORTER string

const (
	SORT_BY_NAME     DISH_SORTER = "name"
	SORT_BY_PRICE    DISH_SORTER = "price"
	SORT_BY_CREATION DISH_SORTER = "createdat"
)

type DishesRepo struct {
	Db *pgx.Conn
}

func NewDishesRepo(connection *pgx.Conn) *DishesRepo {
	return &DishesRepo{Db: connection}
}

func (self *DishesRepo) Create(ctx context.Context, dish *models.Dish) error {
	query := `insert into dishes (id, name, price, created_at)
		values ($1, $2, $3, $4)`
	timestamp := time.Now()
	_, err := self.Db.Exec(ctx, query,
		dish.Id,
		dish.Name,
		dish.Price,
		timestamp)
	if err == nil {
		dish.CreatedAt = &timestamp
	}
	return err
}

func (self *DishesRepo) Update(ctx context.Context, dish *models.Dish) error {
	query := `update dishes set name = $1, price = $2, updated_at=$3`
	timestamp := time.Now()
	_, err := self.Db.Exec(ctx, query,
		dish.Name,
		dish.Price,
		timestamp)
	if err == nil {
		dish.UpdatedAt = &timestamp
	}
	return err
}

func (self *DishesRepo) Delete(ctx context.Context, dish *models.Dish) error {
	query := "update dishes set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := self.Db.Exec(ctx, query, dish.Id, timestamp)
	if err == nil {
		dish.DeletedAt = &timestamp
	}
	return err
}

func (self *DishesRepo) FindByName(ctx context.Context, name string) ([]*models.Dish, error) {
	query := `select * from dishes where name ilike $1`
	r, err := self.Db.Query(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByName[models.Dish])
}

func (self *DishesRepo) FindById(ctx context.Context, id string) (*models.Dish, error) {
	query := `select * from dishes id = $1`
	r, err := self.Db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[models.Dish])
}

func (self *DishesRepo) List(ctx context.Context, page, size int, sorter DISH_SORTER) (int, []*models.Dish, error) {
	offset := (page - 1) * size
	where := "deletedat is not null"
	query := fmt.Sprintf(`select * from dishes
		where %s
        order by %s
        limit $2 offset $3;`, where, sorter)
	r, err := self.Db.Query(ctx, query, sorter, size, offset)
	if err != nil {
		return 0, nil, err
	}
	defer r.Close()
	dishes, err := pgx.CollectRows(r, pgx.RowToAddrOfStructByName[models.Dish])
	if err != nil {
		return 0, nil, err
	}
	pages, err := self.getCount(where)
	if err != nil {
		return 0, nil, err
	}
	return pages, dishes, nil
}

func (self *DishesRepo) getCount(where string) (int, error) {
	query := fmt.Sprintf("select count(*) as pages from dishes where %s", where)
	pages := 0
	err := self.Db.QueryRow(context.Background(), query).Scan(&pages)
	return pages, err
}
