package repositories

import (
	"context"
	"example/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type AddressesRepo struct {
	Db *pgx.Conn
}

func (self *AddressesRepo) Create(address *models.Address) {
	query := `insert into addresses (id, street, house, door, floor)
		values (@id, @street, @house, @door, @floor);`
	self.Db.Exec(context.Background(), query,
		&pgx.NamedArgs{
			"id":     address.Id,
			"street": address.Street,
			"house":  address.House,
			"door":   address.Door,
			"floor":  address.Floor,
		})
}

func (self *AddressesRepo) Update(address *models.Address) error {
	query := `update addresses set
		street = @street,
		house = @house,
		door = @door,
		floor = @floor,
		updatedAt = @now
		where id = @id`
	timestamp := time.Now()
	_, err := self.Db.Exec(context.Background(), query, &pgx.NamedArgs{
		"street":    address.Street,
		"house":     address.House,
		"door":      address.Door,
		"floor":     address.Floor,
		"updatedAt": timestamp,
	})
	if err == nil {
		address.UpdatedAt = &timestamp
	}
	return err
}

func (self *AddressesRepo) Delete(ctx context.Context, address *models.Address) error {
	query := "update addresses set deletedAt = $1 where id = $2"
	timestamp := time.Now()
	_, err := self.Db.Exec(ctx, query, timestamp, address.Id)
	if err == nil {
		address.DeletedAt = &timestamp
	}
	return err
}

func (self *AddressesRepo) FindById(ctx context.Context, id string) (*models.Address, error) {
	query := "select * from addresses where id = $1"
	r, err := self.Db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[models.Address])
}

func (self *AddressesRepo) FindByAddress(ctx context.Context, street, home string) ([]*models.Address, error) {
	query := "select * from addresses where street = $1 and home = $2"
	r, err := self.Db.Query(ctx, query, street, home)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByName[models.Address])
}
