package order

type Order struct {
	ID          int64  `db:"id"`
	UserID      string `db:"user_id"`
	ProductID   int64  `db:"product_id"`
	FlashSaleID int64  `db:"flash_sale_id"`
	Price       int    `db:"price"`
	Status      string `db:"status"`
	CreatedAt   string `db:"created_at"`
}
