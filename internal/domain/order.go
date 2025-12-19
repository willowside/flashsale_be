package domain

import "time"

type Order struct {
	ID          int64      `db:"id"`
	OrderNo     string     `db:"order_no"` // Added
	UserID      string     `db:"user_id"`
	ProductID   int64      `db:"product_id"`
	FlashSaleID int64      `db:"flash_sale_id"`
	Price       int        `db:"price"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	PaidAt      *time.Time `db:"paid_at"`     // Use pointer for nullable columns
	CanceledAt  *time.Time `db:"canceled_at"` // Use pointer for nullable columns
}

type OrderMessage struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Timestamp int64  `json:"timestamp"`
}
