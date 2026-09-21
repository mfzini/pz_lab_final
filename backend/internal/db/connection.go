package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	once       sync.Once
	connection *sql.DB
)

func connect() {
	once.Do(func() {
		cfg := mysql.NewConfig()
		cfg.ParseTime = true
		cfg.User = os.Getenv("MARIADB_USER")
		cfg.Passwd = os.Getenv("MARIADB_PASSWORD")
		cfg.DBName = os.Getenv("MARIADB_DATABASE")
		cfg.Net = "tcp"
		cfg.Addr = fmt.Sprint(os.Getenv("MARIADB_HOST"), ":", os.Getenv("MARIADB_PORT"))
		var err error
		connection, err = sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			panic(err)
		}

		connection.SetConnMaxLifetime(time.Minute * 3)
		connection.SetMaxOpenConns(10)
		connection.SetMaxIdleConns(10)

		pingErr := connection.Ping()
		if pingErr != nil {
			log.Fatal(pingErr)
		}
	})
}

func PrepareStatement(query string) (*sql.Stmt, error) {
	if connection == nil {
		connect()
	}
	stmt, err := connection.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}
