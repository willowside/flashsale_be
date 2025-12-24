package dto

type DLQMessage struct {
	OrderNo string        `json:"order_no"`
	Reason  string        `json:"reason"`
	Payload QueueOrderReq `json:"payload"`
}
