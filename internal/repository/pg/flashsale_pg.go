package pg

import (
	"context"
	"flashsale/internal/domain/flashsale"

	"github.com/jmoiron/sqlx"
)

type FlashSalePG struct {
	db *sqlx.DB
}

func NewFlashSalePG(db *sqlx.DB) *FlashSalePG {
	return &FlashSalePG{db: db}
}

func (r *FlashSalePG) GetActiveSale(ctx context.Context, productID int64) (*flashsale.FlashSale, error) {
	var fs flashsale.FlashSale
	err := r.db.GetContext(ctx, &fs,
		`SELECT * FROM flash_sales WHERE product_id=$1 AND start_time <= NOW() AND end_time >= NOW()`,
		productID,
	)
	return &fs, err
}
