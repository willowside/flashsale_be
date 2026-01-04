package postgres

import (
	"context"
	"flashsale/internal/repository/repositoryiface"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StockPGRepo struct {
	db *pgxpool.Pool
}

func NewStockPGRepo(db *pgxpool.Pool) repositoryiface.StockRepository {
	return &StockPGRepo{
		db: db,
	}
}

func (p *StockPGRepo) GetStock(ctx context.Context, productID int64) (int64, error) {
	var stock int64
	err := p.db.QueryRow(ctx, `SELECT sale_stock FROM flash_sale_products WHERE product_id=$1`, productID).Scan(&stock)
	if err != nil {
		return 0, fmt.Errorf("db GetStock: %w", err)
	}
	return stock, nil
}
