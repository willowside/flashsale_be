package serviceiface

import "context"

/*

	INTERFACE

*/

// OrderPublisher interface
type OrderPublisher interface {
	PublishOrder(ctx context.Context, orderID, userID, productID string) error
}
