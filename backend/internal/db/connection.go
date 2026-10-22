package db

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once sync.Once
	pool *pgxpool.Pool
)

// GetConnection devuelve el pool de conexiones a la DB. Es un pool (y no una
// única *pgx.Conn) porque el server atiende requests en paralelo: una sola
// conexión no admite dos queries al mismo tiempo y tira "conn busy".
func GetConnection() *pgxpool.Pool {
	once.Do(func() {
		var err error
		pool, err = pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB")))
		if err != nil {
			panic(err)
		}
		// Mismo fail-fast que antes: si la DB no responde, no arranca.
		if err := pool.Ping(context.Background()); err != nil {
			panic(err)
		}
	})

	return pool
}
