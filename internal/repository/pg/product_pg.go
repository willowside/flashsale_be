package pg

import (
	"context"
	"flashsale/internal/domain/product"

	"github.com/jmoiron/sqlx"
)

type ProductPG struct {
	db *sqlx.DB
}

func NewProductPG(db *sqlx.DB) *ProductPG {
	return &ProductPG{db: db}
}

func (r *ProductPG) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	var p product.Product
	err := r.db.GetContext(ctx, &p, `SELECT * FROM products WHERE id=$1`, id)
	return &p, err
}

func (r *ProductPG) DecreaseStock(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock = stock - 1 WHERE id=$1 AND stock > 0`,
		id,
	)
	return err
}
