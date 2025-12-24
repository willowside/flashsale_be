package dto

import "time"

type OrderResult struct {
	OrderID         string
	Status          string
	ProcessingState string
	FailReason      *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
