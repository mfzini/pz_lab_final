package db

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
)

var (
	once       sync.Once
	connection *pgx.Conn
)

func GetConnection() *pgx.Conn {
	once.Do(func() {
		var err error
		connection, err = pgx.Connect(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB")))
		if err != nil {
			panic(err)
		}
	})

	return connection
}
