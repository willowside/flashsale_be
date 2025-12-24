package dto

type QueueOrderReq struct {
	OrderNo   string `json:"order_no"`
	UserID    string `json:"user_id"`
	ProductID int64  `json:"product_id"`
}
