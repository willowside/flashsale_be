package pg

import (
	"context"
	"flashsale/internal/domain/order"
	"flashsale/internal/repository"
)

type OrderPG struct{}

func NewOrderPG() *OrderPG {
	return &OrderPG{}
}

func (r *OrderPG) CreateTx(ctx context.Context, tx repository.Tx, o *order.Order) error {
	query := `
        INSERT INTO orders
        (user_id, product_id, flash_sale_id, price, status)
        VALUES ($1, $2, $3, $4, $5)
    `
	return tx.Exec(ctx, query,
		o.UserID, o.ProductID, o.FlashSaleID, o.Price, o.Status,
	)
}
