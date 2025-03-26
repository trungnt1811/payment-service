package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type paymentOrderRepository struct {
	db *gorm.DB
}

func NewPaymentOrderRepository(db *gorm.DB) repotypes.PaymentOrderRepository {
	return &paymentOrderRepository{
		db: db,
	}
}

func (r *paymentOrderRepository) CreatePaymentOrders(
	tx *gorm.DB, ctx context.Context, orders []entities.PaymentOrder, vendorID string,
) ([]entities.PaymentOrder, error) {
	// Validate input
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders to create")
	}

	// Add the vendorID to each order
	for i := range orders {
		orders[i].VendorID = vendorID
	}

	// Insert the payment orders in the current transaction
	err := tx.WithContext(ctx).Create(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders: %w", err)
	}

	// Log the success
	logger.GetLogger().Infof("Successfully created %d payment orders for vendor: %s", len(orders), vendorID)
	return orders, nil
}

// GetActivePaymentOrders retrieves active orders that have not expired or are in "Processing" status.
func (r *paymentOrderRepository) GetActivePaymentOrders(ctx context.Context, network *string) ([]entities.PaymentOrder, error) {
	var orders []entities.PaymentOrder
	currentTime := time.Now().UTC() // Get the current UTC time

	query := r.db.WithContext(ctx).
		Joins("JOIN payment_wallet ON payment_wallet.id = payment_order.wallet_id").                          // Join PaymentWallet with PaymentOrder.
		Preload("Wallet").                                                                                    // Preload the associated Wallet
		Where("(payment_order.status IN (?) AND payment_order.expired_time > ?) OR payment_order.status = ?", // Differentiate logic for `Processing`.
			[]string{constants.Pending, constants.Partial}, // Non-expired statuses.
			currentTime,
			constants.Processing, // Include all `Processing` orders.
		)

	// Filter by network if specified
	if network != nil {
		query = query.Where("payment_order.network = ?", *network)
	}

	// Execute the query and retrieve results
	if err := query.Order("payment_order.expired_time ASC").Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve active orders: %w", err)
	}

	return orders, nil
}

func (r *paymentOrderRepository) UpdatePaymentOrder(
	ctx context.Context,
	orderID uint64,
	updateFunc func(order *entities.PaymentOrder) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order entities.PaymentOrder

		// Retrieve the order with row-level locking
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Wallet").
			Preload("PaymentEventHistories").
			First(&order, "id = ?", orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("payment order id %d not found: %w", orderID, err)
			}
			return fmt.Errorf("failed to retrieve payment order: %w", err)
		}

		// Allow caller to update fields safely within transaction
		if err := updateFunc(&order); err != nil {
			return err
		}

		// Construct update fields dynamically
		updates := map[string]any{
			"status":       order.Status,
			"transferred":  order.Transferred,
			"succeeded_at": order.SucceededAt,
		}
		if order.BlockHeight != 0 {
			updates["block_height"] = order.BlockHeight
		}
		if order.UpcomingBlockHeight != 0 {
			updates["upcoming_block_height"] = order.UpcomingBlockHeight
		}

		// Update the order directly
		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update payment order: %w", err)
		}

		// Determine wallet `in_use` status based on updated order status
		walletInUse := !(order.Status == constants.Success || order.Status == constants.Failed)

		// Update associated wallet's `in_use` status
		if err := tx.Model(&entities.PaymentWallet{}).
			Where("id = ?", order.WalletID).
			Update("in_use", walletInUse).Error; err != nil {
			return fmt.Errorf("failed to update wallet in_use status: %w", err)
		}

		return nil
	})
}

// UpdateOrderToSuccessAndReleaseWallet updates the status of a payment order to "Success" and releases the associated wallet.
func (r *paymentOrderRepository) UpdateOrderToSuccessAndReleaseWallet(
	ctx context.Context,
	orderID uint64,
	succeededAt time.Time,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order status and succeeded_at with row-level locking
		result := tx.Model(&entities.PaymentOrder{}).
			Where("id = ?", orderID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Updates(map[string]any{
				"status":       constants.Success,
				"succeeded_at": succeededAt,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to set order status SUCCESS: %w", result.Error)
		}

		if result.RowsAffected != 1 {
			return fmt.Errorf("unexpected number of rows affected updating payment_order ID %d: %d", orderID, result.RowsAffected)
		}

		// Release wallet (set in_use = false)
		resultWallet := tx.Model(&entities.PaymentWallet{}).
			Joins("JOIN payment_orders ON payment_orders.wallet_id = payment_wallets.id").
			Where("payment_orders.id = ?", orderID).
			Update("in_use", false)

		if resultWallet.Error != nil {
			return fmt.Errorf("failed to release wallet: %w", resultWallet.Error)
		}

		if resultWallet.RowsAffected != 1 {
			return fmt.Errorf("unexpected number of rows affected releasing wallet for payment_order ID %d: %d", orderID, resultWallet.RowsAffected)
		}

		return nil
	})
}

// UpdateOrderNetwork updates the network and block height of a payment order by its request ID.
func (r *paymentOrderRepository) UpdateOrderNetwork(
	ctx context.Context, requestID, network string, blockHeight uint64,
) error {
	// Prepare the update map
	updates := map[string]any{
		"network":      network,
		"block_height": blockHeight,
	}

	// Execute the update
	result := r.db.WithContext(ctx).
		Model(&entities.PaymentOrder{}).
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
func (r *paymentOrderRepository) BatchUpdateOrdersToExpired(ctx context.Context, orderIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Explicit row locking
		var orders []entities.PaymentOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", orderIDs).
			Find(&orders).Error; err != nil {
			return fmt.Errorf("failed to lock orders: %w", err)
		}

		if len(orders) == 0 {
			return fmt.Errorf("no orders found with provided IDs")
		}

		// Perform the update
		result := tx.Model(&entities.PaymentOrder{}).
			Where("id IN ?", orderIDs).
			Update("status", constants.Expired)

		if result.Error != nil {
			return fmt.Errorf("failed to update orders: %w", result.Error)
		}

		return nil
	})
}

// BatchUpdateOrderBlockHeights updates the block heights of multiple payment orders by their OrderIDs.
func (r *paymentOrderRepository) BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error {
	// Check if the lengths of orderIDs and blockHeights match
	if len(orderIDs) != len(blockHeights) {
		return fmt.Errorf("the number of order IDs and block heights must be the same")
	}

	// Build the SQL CASE statement for updating different block heights based on order IDs
	caseSQL := constants.SQLCase
	for i, orderID := range orderIDs {
		caseSQL += fmt.Sprintf(" WHEN id = %d THEN %d", orderID, blockHeights[i])
	}
	caseSQL += constants.SQLEnd

	// Perform the batch update using a single query with CASE and IN
	result := r.db.WithContext(ctx).
		Model(&entities.PaymentOrder{}).
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
func (r *paymentOrderRepository) GetExpiredPaymentOrders(ctx context.Context, network string) ([]entities.PaymentOrder, error) {
	var orders []entities.PaymentOrder

	orderCutoffTime := conf.GetOrderCutoffTime()

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
func (r *paymentOrderRepository) UpdateExpiredOrdersToFailed(ctx context.Context) ([]uint64, error) {
	orderCutoffTime := conf.GetOrderCutoffTime()
	cutoffTime := time.Now().UTC().Add(-orderCutoffTime)

	var allUpdatedIDs []uint64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		offset := 0

		for {
			var orderIDs []uint64

			// Step 1: Select a batch of expired orders with row-level locks
			if err := tx.Model(&entities.PaymentOrder{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("status NOT IN (?)", []string{constants.Success, constants.Failed, constants.Processing}).
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
			if err := tx.Model(&entities.PaymentOrder{}).
				Where("id IN ?", orderIDs).
				Updates(map[string]any{
					"status": constants.Failed,
				}).Error; err != nil {
				return fmt.Errorf("failed to update expired orders: %w", err)
			}

			// Step 3: Update associated wallets to "in_use = false"
			if err := tx.Model(&entities.PaymentWallet{}).
				Where("id IN (?)", tx.Model(&entities.PaymentOrder{}).
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
func (r *paymentOrderRepository) UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error) {
	orderCutoffTime := conf.GetOrderCutoffTime()

	currentTime := time.Now().UTC()
	cutoffTime := currentTime.Add(-orderCutoffTime)

	var updatedIDs []uint64

	// Use GORM transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Fetch the IDs of orders to be updated
		if err := tx.Model(&entities.PaymentOrder{}).
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
		if err := tx.Model(&entities.PaymentOrder{}).
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

func (r *paymentOrderRepository) GetPaymentOrders(
	ctx context.Context,
	limit, offset int,
	vendorID string,
	requestIDs []string,
	status, orderBy, fromAddress, network *string,
	orderDirection constants.OrderDirection,
	startTime, endTime *time.Time,
	timeFilterField *string,
) ([]entities.PaymentOrder, error) {
	var orders []entities.PaymentOrder

	orderColumn := "id" // Default values for ordering
	if orderBy != nil && *orderBy != "" {
		orderColumn = *orderBy
	}

	orderDir := constants.Asc.String() // Default direction
	if orderDirection == constants.Desc {
		orderDir = constants.Desc.String()
	}

	query := r.db.WithContext(ctx).
		Where("vendor_id = ?", vendorID). // Always filter by vendorID
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", orderColumn, orderDir))

	if len(requestIDs) > 0 {
		query = query.Where("request_id IN ?", requestIDs)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if fromAddress != nil {
		query = query.Where("id IN (?)",
			r.db.Table("payment_event_history").
				Select("payment_order_id").
				Where("from_address = ?", *fromAddress),
		)
	}

	if network != nil {
		query = query.Where("network = ?", *network)
	}

	if timeFilterField != nil && (startTime != nil || endTime != nil) {
		timeColumn := *timeFilterField
		if startTime != nil {
			query = query.Where(fmt.Sprintf("%s >= ?", timeColumn), *startTime)
		}
		if endTime != nil {
			query = query.Where(fmt.Sprintf("%s <= ?", timeColumn), *endTime)
		}
	}

	// Execute query
	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payment orders: %w", err)
	}

	return orders, nil
}

// GetPaymentOrdersByID retrieves a single payment order by its ID.
func (r *paymentOrderRepository) GetPaymentOrderByID(ctx context.Context, id uint64) (*entities.PaymentOrder, error) {
	var order entities.PaymentOrder

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
func (r *paymentOrderRepository) GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]entities.PaymentOrder, error) {
	var orders []entities.PaymentOrder

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
func (r *paymentOrderRepository) GetPaymentOrderByRequestID(ctx context.Context, requestID string) (*entities.PaymentOrder, error) {
	var order entities.PaymentOrder

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
func (r *paymentOrderRepository) GetPaymentOrderIDByRequestID(ctx context.Context, requestID string) (uint64, error) {
	var orderID uint64

	// Query to fetch only the ID
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentOrder{}).
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

// ReleaseWalletsForSuccessfulOrders releases wallets that are still marked as in_use for successful orders.
func (r *paymentOrderRepository) ReleaseWalletsForSuccessfulOrders(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&entities.PaymentWallet{}).Where("in_use = true").
		Where("NOT EXISTS (?)",
			r.db.Model(&entities.PaymentOrder{}).
				Select("1").
				Where("payment_order.wallet_id = payment_wallet.id").
				Where("status IN ?", []string{
					constants.Processing,
					constants.Pending,
					constants.Partial,
					constants.Expired,
				}),
		).
		Update("in_use", false).Error
}
