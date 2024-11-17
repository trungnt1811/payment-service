package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentOrderRepository struct {
	db *gorm.DB
}

func NewPaymentOrderRepository(db *gorm.DB) interfaces.PaymentOrderRepository {
	return &paymentOrderRepository{
		db: db,
	}
}

func (r *paymentOrderRepository) CreatePaymentOrders(ctx context.Context, orders []domain.PaymentOrder) ([]domain.PaymentOrder, error) {
	err := r.db.WithContext(ctx).Create(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders: %w", err)
	}
	return orders, nil
}

// GetActivePaymentOrders retrieves active orders that have not expired on a specific network.
func (r *paymentOrderRepository) GetActivePaymentOrders(ctx context.Context, limit, offset int, network string) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder
	currentTime := time.Now().UTC() // Calculate current time in Go

	if err := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id"). // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                           // Preload the associated Wallet
		Limit(limit).
		Offset(offset).
		Where("payment_order.network = ? AND payment_order.status IN (?) AND payment_order.expired_time > ?",
			network,
			[]string{constants.Pending, constants.Partial},
			currentTime,
		).
		Order("payment_order.expired_time ASC"). // Order results by expiration time.
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve pending orders: %w", err)
	}

	return orders, nil
}

// UpdatePaymentOrder updates the status and transferred amount for a payment order,
// and also updates the associated wallet's "in use" status within a transaction to ensure consistency.
func (r *paymentOrderRepository) UpdatePaymentOrder(
	ctx context.Context,
	orderID uint64,
	status, transferredAmount string,
	walletStatus bool,
	blockHeight uint64,
) error {
	// Start a transaction to ensure consistency across both payment order and wallet updates.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Define the fields to be updated for the payment order.
		updateFields := map[string]interface{}{
			"status":       status,
			"transferred":  transferredAmount,
			"block_height": blockHeight,
		}

		// If status is "Success", add the succeeded_at timestamp
		if status == constants.Success {
			updateFields["succeeded_at"] = time.Now().UTC()
		}

		// Update the payment order's status and transferred amount.
		if err := tx.Model(&domain.PaymentOrder{}).
			Where("id = ?", orderID).
			Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update payment order id %d: %w", orderID, err)
		}

		// Fetch the payment order to get the associated wallet ID.
		var paymentOrder domain.PaymentOrder
		if err := tx.First(&paymentOrder, orderID).Error; err != nil {
			return fmt.Errorf("failed to retrieve payment order with id %d: %w", orderID, err)
		}

		// Update the "in_use" status of the associated wallet.
		if err := tx.Model(&domain.PaymentWallet{}).
			Where("id = ?", paymentOrder.WalletID).
			Update("in_use", walletStatus).Error; err != nil {
			return fmt.Errorf("failed to update wallet status for wallet id %d: %w", paymentOrder.WalletID, err)
		}

		// Return nil to commit the transaction.
		return nil
	})
}

// BatchUpdateOrderStatuses updates the statuses of multiple payment orders by their OrderIDs.
func (r *paymentOrderRepository) BatchUpdateOrderStatuses(ctx context.Context, orderIDs []uint64, newStatuses []string) error {
	// Check if the lengths of orderIDs and newStatuses match
	if len(orderIDs) != len(newStatuses) {
		return fmt.Errorf("the number of order IDs and statuses must be the same")
	}

	// Build the SQL CASE statement for updating different statuses based on order IDs
	caseSQL := "CASE"
	for i, orderID := range orderIDs {
		caseSQL += fmt.Sprintf(" WHEN id = %d THEN '%s'::order_status", orderID, newStatuses[i]) // Casting to order_status enum
	}
	caseSQL += " END"

	// Perform the batch update using a single query with CASE and IN
	result := r.db.WithContext(ctx).
		Model(&domain.PaymentOrder{}).
		Where("id IN ?", orderIDs).
		Update("status", gorm.Expr(caseSQL))

	if result.Error != nil {
		return fmt.Errorf("failed to update orders: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no orders found with the provided IDs")
	}

	return nil
}

// BatchUpdateOrderBlockHeights updates the block heights of multiple payment orders by their OrderIDs.
func (r *paymentOrderRepository) BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error {
	// Check if the lengths of orderIDs and blockHeights match
	if len(orderIDs) != len(blockHeights) {
		return fmt.Errorf("the number of order IDs and block heights must be the same")
	}

	// Build the SQL CASE statement for updating different block heights based on order IDs
	caseSQL := "CASE"
	for i, orderID := range orderIDs {
		caseSQL += fmt.Sprintf(" WHEN id = %d THEN %d", orderID, blockHeights[i])
	}
	caseSQL += " END"

	// Perform the batch update using a single query with CASE and IN
	result := r.db.WithContext(ctx).
		Model(&domain.PaymentOrder{}).
		Where("id IN ?", orderIDs).
		Update("block_height", gorm.Expr(caseSQL))

	if result.Error != nil {
		return fmt.Errorf("failed to update block heights: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no orders found with the provided IDs")
	}

	return nil
}

// GetExpiredPaymentOrders retrieves orders for a specific network that are expired within a day.
func (r *paymentOrderRepository) GetExpiredPaymentOrders(ctx context.Context, network string) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder

	// Calculate the time range for the past day
	now := time.Now().UTC()
	cutoffTime := now.Add(-constants.OrderCutoffTime)

	// Execute the query
	if err := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id"). // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                           // Preload the associated Wallet.
		Where("payment_order.network = ? AND payment_order.status NOT IN (?) AND payment_order.expired_time <= ? AND payment_order.expired_time > ?",
			network, []string{constants.Success, constants.Failed}, now, cutoffTime).
		Order("payment_order.block_height ASC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve expired orders: %w", err)
	}

	return orders, nil
}

// UpdateExpiredOrdersToFailed updates all expired orders (longer than 1 day) to "Failed"
// and sets their associated wallets' "in_use" status to false.
func (r *paymentOrderRepository) UpdateExpiredOrdersToFailed(ctx context.Context) error {
	// Calculate the cutoff time (1 day before the current time)
	cutoffTime := time.Now().UTC().Add(-constants.OrderCutoffTime)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Update the status of expired payment orders to "Failed"
		if err := tx.Model(&domain.PaymentOrder{}).
			Where("status = ?", constants.Expired).
			Where("expired_time <= ?", cutoffTime).
			Update("status", constants.Failed).Error; err != nil {
			return fmt.Errorf("failed to update expired orders: %w", err)
		}

		// Step 2: Update the associated wallets' in_use status to false
		if err := tx.Model(&domain.PaymentWallet{}).
			Where("id IN (?)", tx.Model(&domain.PaymentOrder{}).
				Select("wallet_id").
				Where("status = ?", constants.Failed).
				Where("expired_time <= ?", cutoffTime)).
			Update("in_use", false).Error; err != nil {
			return fmt.Errorf("failed to update associated wallets: %w", err)
		}

		return nil
	})
}

// UpdateActiveOrdersToExpired updates all active orders to "Expired"
func (r *paymentOrderRepository) UpdateActiveOrdersToExpired(ctx context.Context) error {
	currentTime := time.Now().UTC()
	// Calculate the cutoff time (1 day before the current time)
	cutoffTime := currentTime.Add(-constants.OrderCutoffTime)
	result := r.db.WithContext(ctx).
		Model(&domain.PaymentOrder{}).
		Where("status IN (?)", []string{constants.Pending, constants.Partial}).
		Where("expired_time > ? AND expired_time <= ?", cutoffTime, currentTime).
		Update("status", constants.Expired)

	if result.Error != nil {
		return fmt.Errorf("failed to update active orders: %w", result.Error)
	}

	return nil
}

// GetPaymentOrderHistories retrieves payment orders by request IDs and optionally filters by status.
func (r *paymentOrderRepository) GetPaymentOrderHistories(
	ctx context.Context,
	limit, offset int,
	requestIDs []string,
	status *string,
) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// Apply filter for request IDs if provided
	if len(requestIDs) > 0 {
		query = query.Preload("PaymentEventHistories").Where("request_id IN ?", requestIDs)
	} else {
		query = query.Preload("PaymentEventHistories")
	}

	// If a status filter is provided, apply it to the query
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Execute query
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment order histories: %w", err)
	}

	return orders, nil
}
