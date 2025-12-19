package mq

type QueueReq struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
}
