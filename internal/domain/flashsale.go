package domain

import "time"

type FlashSaleStatus string

const (
	StatusScheduled FlashSaleStatus = "scheduled"
	StatusActive    FlashSaleStatus = "active"
	StatusEnded     FlashSaleStatus = "ended"
)

type FlashSale struct {
	ID        int64
	Name      string
	StartAt   time.Time
	EndAt     time.Time
	Status    FlashSaleStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsInWindow strictly checks the clock
func (fs *FlashSale) IsInWindow(now time.Time) bool {
	return now.After(fs.StartAt) && now.Before(fs.EndAt)
}

// IsActive checks both status and time
func (fs *FlashSale) IsActive(now time.Time) bool {
	if fs.Status != StatusActive && fs.Status != StatusScheduled {
		return false
	}
	return fs.IsInWindow(now)
}

// TTL calculates Redis expiration with a safety buffer
func (fs *FlashSale) TTL(now time.Time) time.Duration {
	if now.After(fs.EndAt) {
		return 0
	}
	// EndAt - now + 1 hour buffer for late-processing queue messages
	return fs.EndAt.Sub(now) + (60 * time.Minute)
}
