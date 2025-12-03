package db

import (
	"os"

	"github.com/jmoiron/sqlx"
)

func InitPostgres() (*sqlx.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
