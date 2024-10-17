package payment_order

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type paymentOrderRepository struct {
	db *gorm.DB
}

func NewPaymentOrderRepository(db *gorm.DB) interfaces.PaymentOrderRepository {
	return &paymentOrderRepository{
		db: db,
	}
}

func (r *paymentOrderRepository) CreatePaymentOrders(ctx context.Context, orders []model.PaymentOrder) ([]model.PaymentOrder, error) {
	err := r.db.WithContext(ctx).Create(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders: %w", err)
	}
	return orders, nil
}

// GetPendingPaymentOrders retrieves "Pending" orders that have not expired.
func (r *paymentOrderRepository) GetPendingPaymentOrders(ctx context.Context, limit, offset int) ([]model.PaymentOrder, error) {
	var orders []model.PaymentOrder
	currentTime := time.Now().UTC() // Calculate current time in Go

	if err := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id"). // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                           // Preload the associated Wallet
		Limit(limit).
		Offset(offset).
		Where("payment_order.status = ? AND payment_order.expired_time > ?", constants.Pending, currentTime). // Use Go-calculated currentTime
		Order("payment_order.expired_time ASC").                                                              // Order results by expiration time.
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

		// Update the payment order's status and transferred amount.
		if err := tx.Model(&model.PaymentOrder{}).
			Where("id = ?", orderID).
			Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update payment order id %d: %w", orderID, err)
		}

		// Fetch the payment order to get the associated wallet ID.
		var paymentOrder model.PaymentOrder
		if err := tx.First(&paymentOrder, orderID).Error; err != nil {
			return fmt.Errorf("failed to retrieve payment order with id %d: %w", orderID, err)
		}

		// Update the "in_use" status of the associated wallet.
		if err := tx.Model(&model.PaymentWallet{}).
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
		Model(&model.PaymentOrder{}).
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

// GetExpiredPaymentOrders retrieves orders that are expired within a day.
func (r *paymentOrderRepository) GetExpiredPaymentOrders(ctx context.Context) ([]model.PaymentOrder, error) {
	var orders []model.PaymentOrder

	// Calculate the time range for the past day
	now := time.Now().UTC()
	cutoffTime := time.Now().Add(-constants.OrderCutoffTime)

	// Execute the query
	if err := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id"). // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                           // Preload the associated Wallet.
		Where("payment_order.status != ? AND payment_order.expired_time <= ? AND payment_order.expired_time > ?",
			constants.Success, now, cutoffTime).
		Order("payment_order.block_height ASC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve expired orders: %w", err)
	}

	return orders, nil
}

// UpdateExpiredOrdersToFailed updates all expired orders (longer than 1 day) to "FAILED"
// and sets their associated wallets' "in_use" status to false.
func (r *paymentOrderRepository) UpdateExpiredOrdersToFailed(ctx context.Context) error {
	// Calculate the cutoff time in Go (1 day before the current time)
	cutoffTime := time.Now().UTC().Add(-constants.OrderCutoffTime)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Update the status of expired payment orders to "FAILED"
		if err := tx.Model(&model.PaymentOrder{}).
			Where("status = ?", constants.Expired).
			Where("expired_time <= ?", cutoffTime).
			Update("status", constants.Failed).Error; err != nil {
			return fmt.Errorf("failed to update expired orders: %w", err)
		}

		// Step 2: Update the associated wallets' in_use status to false
		if err := tx.Model(&model.PaymentWallet{}).
			Where("id IN (?)", tx.Model(&model.PaymentOrder{}).
				Select("wallet_id").
				Where("status = ?", constants.Failed).
				Where("expired_time <= ?", cutoffTime)).
			Update("in_use", false).Error; err != nil {
			return fmt.Errorf("failed to update associated wallets: %w", err)
		}

		return nil
	})
}
