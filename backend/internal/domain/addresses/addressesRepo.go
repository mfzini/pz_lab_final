package addresses

import (
	"context"
	"fmt"
	"os"
	"time"

	"example/internal/domain/common"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Columnas de lectura, siempre iguales. coordinates va con ::text: sin el cast
// el tipo point crudo no se puede escanear en *string ("cannot scan point"),
// y a.eplicar la columna con "a.*" hacía que pgx la viera dos veces y fallara.
const columns = `id, user_id, street, house, door, floor,
	coordinates::text as coordinates, detail, created_at, updated_at`

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repo {
	return &repo{db}
}

func (this *repo) create(ctx context.Context, address *Model) error {
	query := `insert into addresses (id, user_id, street, house, door, floor, coordinates, detail)
		values ($1, $2, $3, $4, $5, $6, $7, $8);`
	_, err := this.db.Exec(ctx, query, address.Id, address.User_ID, address.Street, address.House,
		address.Door, address.Floor, address.Coordinates, address.Detail)
	return err
}

func (this *repo) update(ctx context.Context, address *Model) error {
	query := `update addresses set
		street = @street,
		house = @house,
		door = @door,
		floor = @floor,
		coordinates = @coordinates,
		detail = @detail,
		updated_at = @now
		where id = @id`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, &pgx.NamedArgs{
		"street":      address.Street,
		"house":       address.House,
		"door":        address.Door,
		"floor":       address.Floor,
		"coordinates": address.Coordinates,
		"detail":      address.Detail,
		"id":          address.Id,
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
	_, err := this.db.Exec(ctx, query, timestamp, address.Id)
	if err == nil {
		address.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindById(ctx context.Context, id string) (*Model, error) {
	query := `select ` + columns + `
		from addresses where id = $1 and deleted_at is null`
	r, err := this.db.Query(ctx, query, id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByNameLax[Model])
}

func (this *repo) FindByStreetAndHouse(ctx context.Context, street, house string, page, size int) ([]*Model, int, error) {
	where := "where street = $1 and house = $2 and deleted_at is null"
	query := "select " + columns + " from addresses " + where + `
		order by created_at desc, id
		limit $3 offset $4`
	rows, err := this.db.Query(ctx, query, street, house, size, (page-1)*size)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, 0, err
	}
	defer rows.Close()

	addresses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		return nil, 0, err
	}

	totalRows := 0
	if err := this.db.QueryRow(ctx, "select count(*) from addresses "+where, street, house).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return addresses, common.TotalPagesFor(totalRows, size), nil
}

// FindByUserID lista las direcciones del usuario con filtro opcional de
// calle: street es la búsqueda asíncrona del checkout (ilike, no un select
// con todos los registros).
func (this *repo) FindByUserID(ctx context.Context, id, street string, page, size int) ([]*Model, int, error) {
	where := "where a.user_id = $1 and a.deleted_at is null"
	args := []any{id}
	if street != "" {
		args = append(args, "%"+street+"%")
		where += fmt.Sprintf(" and a.street ilike $%d", len(args))
	}

	args = append(args, size, (page-1)*size)
	query := `select a.id, a.user_id, a.street, a.house, a.door, a.floor,
		a.coordinates::text as coordinates, a.detail, a.created_at, a.updated_at
		from addresses a
		` + where + fmt.Sprintf(`
		order by a.created_at desc, a.id
		limit $%d offset $%d`, len(args)-1, len(args))
	rows, err := this.db.Query(ctx, query, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, 0, err
	}
	defer rows.Close()

	addresses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		return nil, 0, err
	}

	countArgs := args[:len(args)-2]
	totalRows := 0
	countQuery := `select count(*) from addresses a ` + where
	if err := this.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return addresses, common.TotalPagesFor(totalRows, size), nil
}
