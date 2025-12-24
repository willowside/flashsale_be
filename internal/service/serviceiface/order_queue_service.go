package serviceiface

import "context"

type OrderQueueService interface {
	CreateAndQueueOrder(ctx context.Context, userID string, productID string) (string, error)
}
