package addresses

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type repo struct {
	db *pgx.Conn
}

func NewRepo(db *pgx.Conn) *repo {
	return &repo{db}
}

func (this *repo) create(ctx context.Context, address *Model) error {
	query := `insert into addresses (id, user_id, street, house, door, floor, coordinates)
		values ($1, $2, $3, $4, $5, $6, $7);`
	_, err := this.db.Exec(ctx, query, address.ID, address.User_ID, address.Street, address.House, address.Door, address.Floor, address.Coordinates)
	return err
}

func (this *repo) update(ctx context.Context, address *Model) error {
	query := `update addresses set
		street = @street,
		house = @house,
		door = @door,
		floor = @floor,
		coordinates = @coordinates
		updated_at = @now
		where id = @id`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, &pgx.NamedArgs{
		"street":      address.Street,
		"house":       address.House,
		"door":        address.Door,
		"floor":       address.Floor,
		"coordinates": address.Coordinates,
		"id":          address.ID,
		"now":         timestamp,
	})
	if err == nil {
		address.UpdatedAt = &timestamp
	}
	return err
}

func (this *repo) delete(ctx context.Context, address *Model) error {
	query := "update addresses set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := this.db.Exec(ctx, query, timestamp, address.ID)
	if err == nil {
		address.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindById(ctx context.Context, id string) (*Model, error) {
	query := "select id, street, house, coordinates::text as coordinates, floor, door, created_at, updated_at from addresses where id = $1"
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByNameLax[Model])
}

func (this *repo) FindByStreetAndHouse(ctx context.Context, street, house string) ([]*Model, error) {
	query := "select * from addresses where street = $1 and house = $2"
	r, err := this.db.Query(ctx, query, street, house)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) FindByUserID(ctx context.Context, id string) ([]*Model, error) {
	query := `select a.*, a.coordinates::text as coordinates
		from addresses a
		where a.user_id = $1`
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByNameLax[Model])
}
