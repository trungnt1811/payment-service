package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type PaymentOrderRepository struct {
	db *gorm.DB
}

func NewPaymentOrderRepository(db *gorm.DB) *PaymentOrderRepository {
	return &PaymentOrderRepository{
		db: db,
	}
}

func (r *PaymentOrderRepository) CreatePaymentOrders(tx *gorm.DB, ctx context.Context, orders []domain.PaymentOrder) ([]domain.PaymentOrder, error) {
	// Validate input
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders to create")
	}

	// Insert the payment orders in the current transaction
	err := tx.WithContext(ctx).Create(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders: %w", err)
	}

	// Log the success
	logger.GetLogger().Infof("Successfully created %d payment orders", len(orders))
	return orders, nil
}

// GetActivePaymentOrders retrieves active orders that have not expired on a specific network.
func (r *PaymentOrderRepository) GetActivePaymentOrders(ctx context.Context, limit, offset int, network *string) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder
	currentTime := time.Now().UTC() // Calculate current time in Go

	query := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id"). // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                           // Preload the associated Wallet
		Limit(limit).
		Offset(offset).
		Where("payment_order.status IN (?) AND payment_order.expired_time > ?",
			[]string{constants.Pending, constants.Processing, constants.Partial},
			currentTime,
		)

	// Check if the network is specified
	if network != nil {
		query = query.Where("payment_order.network = ?", *network)
	}

	if err := query.Order("payment_order.expired_time ASC"). // Order results by expiration time.
									Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve pending orders: %w", err)
	}

	return orders, nil
}

func (r *PaymentOrderRepository) UpdatePaymentOrder(ctx context.Context, order *domain.PaymentOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Construct the updates map, excluding block_height and upcoming_block_height if they are 0
		updates := map[string]interface{}{
			"status":       order.Status,
			"transferred":  order.Transferred,
			"succeeded_at": order.SucceededAt,
		}
		// Only include block_height if it's not 0
		if order.BlockHeight != 0 {
			updates["block_height"] = order.BlockHeight
		}
		// Only include upcoming_block_height if it's not 0
		if order.UpcomingBlockHeight != 0 {
			updates["upcoming_block_height"] = order.UpcomingBlockHeight
		}
		// Only include network if it's not empty
		if order.Network != "" {
			updates["network"] = order.Network
		}

		// Step 1: Update the payment order with row-level locking
		if err := tx.Model(&domain.PaymentOrder{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update payment order: %w", err)
		}

		// Step 2: Fetch the WalletID with row-level locking
		var paymentOrder struct {
			WalletID uint64
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Model(&domain.PaymentOrder{}).
			Select("wallet_id").
			Where("id = ?", order.ID).
			First(&paymentOrder).Error; err != nil {
			return fmt.Errorf("failed to retrieve payment order with id %d: %w", order.ID, err)
		}

		// Step 3: Determine the wallet status
		walletStatus := true
		if order.Status == constants.Success || order.Status == constants.Failed {
			walletStatus = false
		}

		// Step 4: Update the wallet's `in_use` status
		result := tx.Model(&domain.PaymentWallet{}).
			Where("id = ?", paymentOrder.WalletID).
			Update("in_use", walletStatus)
		if result.Error != nil {
			return fmt.Errorf("failed to update wallet status for wallet id %d: %w", paymentOrder.WalletID, result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("wallet id %d not updated, it might have been modified by another transaction", paymentOrder.WalletID)
		}

		return nil
	})
}

// UpdateOrderNetwork updates the network and block height of a payment order by its request ID.
func (r *PaymentOrderRepository) UpdateOrderNetwork(ctx context.Context, requestID, network string, blockHeight uint64) error {
	// Prepare the update map
	updates := map[string]interface{}{
		"network":      network,
		"block_height": blockHeight,
	}

	// Execute the update
	result := r.db.WithContext(ctx).
		Model(&domain.PaymentOrder{}).
		Where("request_id = ?", requestID).
		Updates(updates)

	// Handle errors from the update query
	if result.Error != nil {
		return fmt.Errorf("failed to update order network and block height: %w", result.Error)
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		return fmt.Errorf("no order found with the provided ID")
	}

	return nil
}

// BatchUpdateOrdersToExpired updates the status of multiple payment orders to "Expired" by their OrderIDs.
func (r *PaymentOrderRepository) BatchUpdateOrdersToExpired(ctx context.Context, orderIDs []uint64) error {
	// Build the SQL CASE statement for updating different statuses based on order IDs
	caseSQL := constants.SqlCase
	for _, orderID := range orderIDs {
		caseSQL += fmt.Sprintf(" WHEN id = %d THEN '%s'::order_status", orderID, constants.Expired)
	}
	caseSQL += constants.SqlEnd

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
func (r *PaymentOrderRepository) BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error {
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
func (r *PaymentOrderRepository) GetExpiredPaymentOrders(ctx context.Context, network string, orderCutoffTime time.Duration) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder

	// Calculate the time range for the past day
	now := time.Now().UTC()
	cutoffTime := now.Add(-orderCutoffTime)

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

// UpdateExpiredOrdersToFailed updates all expired orders to "Failed" and sets their associated wallets' "in_use" status to false.
// It returns the IDs of the updated orders.
func (r *PaymentOrderRepository) UpdateExpiredOrdersToFailed(ctx context.Context, orderCutoffTime time.Duration) ([]uint64, error) {
	cutoffTime := time.Now().UTC().Add(-orderCutoffTime)

	var allUpdatedIDs []uint64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		offset := 0

		for {
			var orderIDs []uint64

			// Step 1: Select a batch of expired orders with row-level locks
			if err := tx.Model(&domain.PaymentOrder{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("status NOT IN (?)", []string{constants.Success, constants.Failed}).
				Where("expired_time <= ?", cutoffTime).
				Limit(constants.BatchSize).
				Offset(offset).
				Pluck("id", &orderIDs).Error; err != nil {
				return fmt.Errorf("failed to fetch expired order IDs: %w", err)
			}

			if len(orderIDs) == 0 {
				break // No more expired orders to process
			}

			// Step 2: Update their status to "Failed"
			if err := tx.Model(&domain.PaymentOrder{}).
				Where("id IN ?", orderIDs).
				Updates(map[string]interface{}{
					"status": constants.Failed,
				}).Error; err != nil {
				return fmt.Errorf("failed to update expired orders: %w", err)
			}

			// Step 3: Update associated wallets to "in_use = false"
			if err := tx.Model(&domain.PaymentWallet{}).
				Where("id IN (?)", tx.Model(&domain.PaymentOrder{}).
					Select("wallet_id").
					Where("id IN ?", orderIDs)).
				Update("in_use", false).Error; err != nil {
				return fmt.Errorf("failed to update associated wallets: %w", err)
			}

			// Append the processed IDs to the result slice
			allUpdatedIDs = append(allUpdatedIDs, orderIDs...)

			offset += constants.BatchSize
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return allUpdatedIDs, nil
}

// UpdateActiveOrdersToExpired updates all active orders to "Expired" and returns their IDs.
func (r *PaymentOrderRepository) UpdateActiveOrdersToExpired(ctx context.Context, orderCutoffTime time.Duration) ([]uint64, error) {
	currentTime := time.Now().UTC()
	cutoffTime := currentTime.Add(-orderCutoffTime)

	var updatedIDs []uint64

	// Use GORM transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Fetch the IDs of orders to be updated
		if err := tx.Model(&domain.PaymentOrder{}).
			Select("id").
			Where("status IN (?)", []string{constants.Pending, constants.Partial}).
			Where("expired_time > ? AND expired_time <= ?", cutoffTime, currentTime).
			Scan(&updatedIDs).Error; err != nil {
			return fmt.Errorf("failed to fetch IDs of active orders to be updated: %w", err)
		}

		if len(updatedIDs) == 0 {
			// No orders to update, return early
			return nil
		}

		// Step 2: Update the status of the fetched orders
		if err := tx.Model(&domain.PaymentOrder{}).
			Where("id IN (?)", updatedIDs).
			Update("status", constants.Expired).Error; err != nil {
			return fmt.Errorf("failed to update active orders: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedIDs, nil
}

// GetPaymentOrders retrieves payment orders by request IDs and optionally filters by status.
func (r *PaymentOrderRepository) GetPaymentOrders(
	ctx context.Context,
	limit, offset int,
	requestIDs []string,
	status, orderBy *string,
	orderDirection constants.OrderDirection,
) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder

	orderColumn := "id" // Default values for ordering
	if orderBy != nil && *orderBy != "" {
		orderColumn = *orderBy
	}

	orderDir := string(constants.Asc) // Default direction
	if orderDirection == constants.Desc {
		orderDir = string(constants.Desc)
	}

	query := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", orderColumn, orderDir))

	// If a status filter is provided, apply it to the query
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Apply filter for request IDs if provided
	if len(requestIDs) > 0 {
		query = query.Where("request_id IN ?", requestIDs)
	}

	// Execute query
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment order histories: %w", err)
	}

	return orders, nil
}

// GetPaymentOrdersByID retrieves a single payment order by its ID.
func (r *PaymentOrderRepository) GetPaymentOrderByID(ctx context.Context, id uint64) (*domain.PaymentOrder, error) {
	var order domain.PaymentOrder

	// Execute query to find the payment order by ID with preloaded PaymentEventHistories
	if err := r.db.WithContext(ctx).
		Preload("Wallet").
		Preload("PaymentEventHistories").
		First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment order with ID %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to retrieve payment order: %w", err)
	}

	return &order, nil
}

// GetPaymentOrdersByIDs retrieves multiple payment orders by their IDs.
func (r *PaymentOrderRepository) GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]domain.PaymentOrder, error) {
	var orders []domain.PaymentOrder

	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided to retrieve payment orders")
	}

	// Execute query to find the payment orders by IDs with preloaded PaymentEventHistories
	if err := r.db.WithContext(ctx).
		Preload("Wallet").
		Preload("PaymentEventHistories").
		Where("id IN ?", ids).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment orders: %w", err)
	}

	return orders, nil
}

// GetPaymentOrderByRequestID retrieves a single payment order by its request ID.
func (r *PaymentOrderRepository) GetPaymentOrderByRequestID(ctx context.Context, requestID string) (*domain.PaymentOrder, error) {
	var order domain.PaymentOrder

	// Execute query to find the payment order by request ID with preloaded PaymentEventHistories
	if err := r.db.WithContext(ctx).
		Preload("Wallet").
		Preload("PaymentEventHistories").
		First(&order, "request_id = ?", requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment order with request ID %s not found: %w", requestID, err)
		}
		return nil, fmt.Errorf("failed to retrieve payment order: %w", err)
	}

	return &order, nil
}

// GetPaymentOrderIDByRequestID retrieves the ID of a payment order by its request ID.
func (r *PaymentOrderRepository) GetPaymentOrderIDByRequestID(ctx context.Context, requestID string) (uint64, error) {
	var orderID uint64

	// Query to fetch only the ID
	if err := r.db.WithContext(ctx).
		Model(&domain.PaymentOrder{}).
		Select("id").
		Where("request_id = ?", requestID).
		Scan(&orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("payment order with request ID %s not found: %w", requestID, err)
		}
		return 0, fmt.Errorf("failed to retrieve payment order ID: %w", err)
	}

	return orderID, nil
}
