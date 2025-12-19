package repository

import (
	"flashsale/internal/repository/postgres"
	"flashsale/internal/repository/repositoryiface"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The factory function's signature must include all external dependencies
// required by any concrete implementation it might create (pool and locker).
func NewOrderRepository(pool *pgxpool.Pool, dbType string) repositoryiface.OrderRepository {

	// The factory logic decides which concrete implementation to return.
	switch dbType {
	case "postgres":
		return postgres.NewOrderPGRepo(pool)

	default:
		// Best to return an error or panic if the type is unknown, but nil works for this example.
		return nil
	}
}

func NewProductRepository(pool *pgxpool.Pool, dbType string) repositoryiface.ProductRepository {

	// The factory logic decides which concrete implementation to return.
	switch dbType {
	case "postgres":
		return postgres.NewProductPGRepo(pool)

	default:
		// Best to return an error or panic if the type is unknown, but nil works for this example.
		return nil
	}
}
