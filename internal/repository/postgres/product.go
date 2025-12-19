package postgres

import (
	"context"
	"flashsale/internal/repository/repositoryiface"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductPGRepo struct {
	db *pgxpool.Pool
}

func NewProductPGRepo(db *pgxpool.Pool) repositoryiface.ProductRepository {
	return &ProductPGRepo{
		db: db,
	}
}

func (p *ProductPGRepo) GetStock(ctx context.Context, productID int64) (int64, error) {
	var stock int64
	err := p.db.QueryRow(ctx, `SELECT stock FROM products WHERE id=$1`, productID).Scan(&stock)
	if err != nil {
		return 0, fmt.Errorf("db GetStock: %w", err)
	}
	return stock, nil
}
