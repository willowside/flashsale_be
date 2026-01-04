package domain

type FlashSaleProduct struct {
	ID          int64
	FlashSaleID int64
	ProductID   int64
	SaleStock   int
	SalePrice   int
}
