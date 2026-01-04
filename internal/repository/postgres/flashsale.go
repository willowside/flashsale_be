package postgres

import (
	"context"
	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FlashSalePGRepo struct {
	pool *pgxpool.Pool
}

func NewFlashSalePGRepo(pool *pgxpool.Pool) repositoryiface.FlashSaleRepository {
	return &FlashSalePGRepo{pool: pool}
}

func (r *FlashSalePGRepo) GetActiveFlashSale(ctx context.Context) (*domain.FlashSale, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, start_at, end_at, status, created_at, updated_at
		FROM flash_sales
		WHERE status = 'active'
		LIMIT 1
	`)

	var fs domain.FlashSale
	err := row.Scan(
		&fs.ID,
		&fs.Name,
		&fs.StartAt,
		&fs.EndAt,
		&fs.Status,
		&fs.CreatedAt,
		&fs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (r *FlashSalePGRepo) GetFlashSaleByID(ctx context.Context, id int64) (*domain.FlashSale, error) {
	row := r.pool.QueryRow(ctx, `
        SELECT id, name, start_at, end_at, status, created_at, updated_at
        FROM flash_sales
        WHERE id = $1
    `, id)

	var fs domain.FlashSale
	err := row.Scan(
		&fs.ID,
		&fs.Name,
		&fs.StartAt,
		&fs.EndAt,
		&fs.Status,
		&fs.CreatedAt,
		&fs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (r *FlashSalePGRepo) GetFlashSaleProducts(ctx context.Context, flashSaleID int64) ([]domain.FlashSaleProduct, error) {

	rows, err := r.pool.Query(ctx, `
		SELECT id, flash_sale_id, product_id, sale_stock, sale_price
		FROM flash_sale_products
		WHERE flash_sale_id = $1
	`, flashSaleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.FlashSaleProduct
	for rows.Next() {
		var p domain.FlashSaleProduct
		if err := rows.Scan(
			&p.ID,
			&p.FlashSaleID,
			&p.ProductID,
			&p.SaleStock,
			&p.SalePrice,
		); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *FlashSalePGRepo) GetFlashSaleProduct(ctx context.Context, flashSaleID int64, productID string) (*domain.FlashSaleProduct, error) {
	row := r.pool.QueryRow(ctx, `
        SELECT id, flash_sale_id, product_id, sale_stock, sale_price
        FROM flash_sale_products
        WHERE flash_sale_id = $1 AND product_id = $2
    `, flashSaleID, productID)

	var p domain.FlashSaleProduct
	err := row.Scan(
		&p.ID,
		&p.FlashSaleID,
		&p.ProductID,
		&p.SaleStock,
		&p.SalePrice,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *FlashSalePGRepo) UpdateStatus(ctx context.Context, id int64, status domain.FlashSaleStatus) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE flash_sales 
        SET status = $2, updated_at = NOW() 
        WHERE id = $1
    `, id, status)
	return err
}
