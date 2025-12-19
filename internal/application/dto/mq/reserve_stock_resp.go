package mq

type QueueResp struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
