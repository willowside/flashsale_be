package flashsale

type FlashSale struct {
	ID        int64  `db:"id"`
	ProductID int64  `db:"product_id"`
	StartTime string `db:"start_time"`
	EndTime   string `db:"end_time"`
	SalePrice int    `db:"sale_price"`
	CreatedAt int    `db:"created_at"`
}
