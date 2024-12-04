package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

const (
	keyPrefixPaymentOrder        = "payment_order_"
	defaultCacheTimePaymentOrder = 5 * time.Second
)

type paymentOrderCache struct {
	paymentOrderRepository *PaymentOrderRepository
	cache                  infrainterfaces.CacheRepository
	config                 *conf.Configuration
}

func NewPaymentOrderCacheRepository(
	repo *PaymentOrderRepository,
	cache infrainterfaces.CacheRepository,
	config *conf.Configuration,
) interfaces.PaymentOrderRepository {
	return &paymentOrderCache{
		paymentOrderRepository: repo,
		cache:                  cache,
		config:                 config,
	}
}

func (c *paymentOrderCache) CreatePaymentOrders(ctx context.Context, orders []domain.PaymentOrder) ([]domain.PaymentOrder, error) {
	// Create orders in the repository (DB)
	createdOrders, err := c.paymentOrderRepository.CreatePaymentOrders(ctx, orders)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders in repository: %w", err)
	}

	// Cache each created order
	for _, order := range createdOrders {
		key := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(order.ID, 10)} // Ensure proper ID conversion
		if cacheErr := c.cache.SaveItem(key, order, c.config.GetExpiredOrderTime()); cacheErr != nil {
			// Log the caching error but don't stop the entire operation
			logger.GetLogger().Warnf("Failed to cache payment order ID %d: %v", order.ID, cacheErr)
		}
	}

	return createdOrders, nil
}

func (c *paymentOrderCache) GetActivePaymentOrders(ctx context.Context, limit, offset int, network *string) ([]domain.PaymentOrder, error) {
	// Handle nil network gracefully in the cache key
	networkStr := "nil"
	if network != nil {
		networkStr = *network
	}

	// Generate a consistent cache key
	key := &caching.Keyer{
		Raw: fmt.Sprintf("%sGetActivePaymentOrders_network:%s_limit:%d_offset:%d",
			keyPrefixPaymentOrder, networkStr, limit, offset),
	}

	// Attempt to retrieve data from cache
	var paymentOrders []domain.PaymentOrder
	if err := c.cache.RetrieveItem(key, &paymentOrders); err == nil {
		// Cache hit: return the cached data
		return paymentOrders, nil
	}

	// Cache miss: fetch from repository
	paymentOrders, err := c.paymentOrderRepository.GetActivePaymentOrders(ctx, limit, offset, network)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active payment orders: %w", err)
	}

	// Store the fetched data in the cache for future use
	if cacheErr := c.cache.SaveItem(key, paymentOrders, defaultCacheTimePaymentOrder); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache active payment orders for network %s: %v", networkStr, cacheErr)
	}

	return paymentOrders, nil
}

func (c *paymentOrderCache) UpdatePaymentOrder(
	ctx context.Context,
	order *domain.PaymentOrder,
) error {
	// Construct the cache key for the payment order
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(order.ID, 10)}

	// Attempt to retrieve the payment order from the cache
	var cachedOrder domain.PaymentOrder
	cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)
	if cacheErr != nil {
		// Log the cache miss but proceed with the update
		logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", order.ID, cacheErr)
	}

	// Update the payment order in the repository (DB)
	if err := c.paymentOrderRepository.UpdatePaymentOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update payment order in repository: %w", err)
	}

	// Update the cache with the new order data
	if cacheErr == nil {
		// Merge changes from the updated order into the cached order
		mergePaymentOrderFields(&cachedOrder, order)

		// Save the updated order back into the cache
		if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
			logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", order.ID, saveErr)
		}
	} else {
		// Cache miss: add the updated order directly to the cache
		if saveErr := c.cache.SaveItem(cacheKey, order, c.config.GetExpiredOrderTime()); saveErr != nil {
			logger.GetLogger().Warnf("Failed to cache updated payment order ID %d: %v", order.ID, saveErr)
		}
	}

	return nil
}

// mergePaymentOrderFields updates only the modified fields of `src` into `dst`.
func mergePaymentOrderFields(dst, src *domain.PaymentOrder) {
	if src.Status != "" {
		dst.Status = src.Status
		if src.Status == constants.Success {
			dst.SucceededAt = src.SucceededAt
		}
	}
	if src.Transferred != "" {
		dst.Transferred = src.Transferred
	}
	if src.Network != "" {
		dst.Network = src.Network
	}
}

func (c *paymentOrderCache) UpdateOrderStatus(
	ctx context.Context,
	orderID uint64,
	newStatus string,
) error {
	// Construct the cache key for the payment order
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

	// Attempt to retrieve the payment order from the cache
	var cachedOrder domain.PaymentOrder
	cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

	// If the order is found in the cache, update it
	if cacheErr == nil {
		// Update the status in the cached order
		cachedOrder.Status = newStatus

		// Save the updated order back into the cache
		if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
			logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", orderID, saveErr)
		}
	} else {
		// Log cache miss but proceed with the update
		logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", orderID, cacheErr)
	}

	// Update the order status in the repository (DB)
	if err := c.paymentOrderRepository.UpdateOrderStatus(ctx, orderID, newStatus); err != nil {
		return fmt.Errorf("failed to update payment order status in repository: %w", err)
	}

	return nil
}

func (c *paymentOrderCache) UpdateOrderNetwork(
	ctx context.Context,
	orderID uint64,
	network string,
) error {
	// Construct the cache key for the payment order
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

	// Attempt to retrieve the payment order from the cache
	var cachedOrder domain.PaymentOrder
	cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

	// If the order is found in the cache, update it
	if cacheErr == nil {
		// Update the network in the cached order
		cachedOrder.Network = network

		// Save the updated order back into the cache
		if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
			logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", orderID, saveErr)
		}
	} else {
		// Log cache miss but proceed with the update
		logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", orderID, cacheErr)
	}

	latestBlock, err := blockchain.GetLatestBlockFromCache(ctx, network, c.cache)
	if err != nil {
		return fmt.Errorf("failed to get latest block from cache: %w", err)
	}

	// Update the order network in the repository (DB)
	if err := c.paymentOrderRepository.UpdateOrderNetwork(ctx, orderID, network, latestBlock); err != nil {
		return fmt.Errorf("failed to update payment order network in repository: %w", err)
	}

	return nil
}

func (c *paymentOrderCache) BatchUpdateOrderStatuses(
	ctx context.Context,
	orderIDs []uint64,
	newStatuses []string,
) error {
	// Ensure that orderIDs and newStatuses have the same length
	if len(orderIDs) != len(newStatuses) {
		return fmt.Errorf("mismatched lengths: orderIDs=%d, newStatuses=%d", len(orderIDs), len(newStatuses))
	}

	// Iterate over the order IDs and new statuses
	for i, orderID := range orderIDs {
		// Construct the cache key
		cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

		// Attempt to retrieve the order from the cache
		var cachedOrder domain.PaymentOrder
		cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

		// If found in cache, update it
		if cacheErr == nil {
			// Update status in the cached order
			cachedOrder.Status = newStatuses[i]

			// Save the updated order back into the cache
			if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
				logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", orderID, saveErr)
			}
		} else {
			// Log the cache miss but continue
			logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", orderID, cacheErr)
		}
	}

	// Batch update the order status in the repository (DB)
	if err := c.paymentOrderRepository.BatchUpdateOrderStatuses(ctx, orderIDs, newStatuses); err != nil {
		return err
	}

	return nil
}

func (c *paymentOrderCache) BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error {
	// Ensure that orderIDs and newStatuses have the same length
	if len(orderIDs) != len(blockHeights) {
		return fmt.Errorf("mismatched lengths: orderIDs=%d, blockHeights=%d", len(orderIDs), len(blockHeights))
	}

	// Iterate over the orderIDs and blockheights
	for i, orderID := range orderIDs {
		// Construct the cache key
		cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

		// Attempt to retrieve the order from the cache
		var cachedOrder domain.PaymentOrder
		cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

		// If found in cache, update it
		if cacheErr == nil {
			// Update block height in the cached order
			cachedOrder.BlockHeight = blockHeights[i]

			// Save the updated order back into the cache
			if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
				logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", orderID, saveErr)
			}
		} else {
			// Log the cache miss but continue
			logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", orderID, cacheErr)
		}
	}

	// Batch update the order block height in the repository (DB)
	if err := c.paymentOrderRepository.BatchUpdateOrderBlockHeights(ctx, orderIDs, blockHeights); err != nil {
		return err
	}

	return nil
}

func (c *paymentOrderCache) GetExpiredPaymentOrders(ctx context.Context, network string) ([]domain.PaymentOrder, error) {
	// Generate a consistent cache key
	key := &caching.Keyer{Raw: fmt.Sprintf("%sGetExpiredPaymentOrders_network:%s", keyPrefixPaymentOrder, network)}

	// Attempt to retrieve data from cache
	var paymentOrders []domain.PaymentOrder
	if err := c.cache.RetrieveItem(key, &paymentOrders); err == nil {
		// Cache hit: return the cached data
		return paymentOrders, nil
	}

	// Cache miss: fetch from repository
	paymentOrders, err := c.paymentOrderRepository.GetExpiredPaymentOrders(ctx, network)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expired payment orders: %w", err)
	}

	// Store the fetched data in the cache for future use
	if cacheErr := c.cache.SaveItem(key, paymentOrders, defaultCacheTimePaymentOrder); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache expired payment orders for network %s: %v", network, cacheErr)
	}

	return paymentOrders, nil
}

func (c *paymentOrderCache) UpdateExpiredOrdersToFailed(ctx context.Context) error {
	return c.paymentOrderRepository.UpdateExpiredOrdersToFailed(ctx)
}

func (c *paymentOrderCache) UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error) {
	// Call repository to update active orders to expired and get the updated IDs
	updatedIDs, err := c.paymentOrderRepository.UpdateActiveOrdersToExpired(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update active orders to expired in repository: %w", err)
	}

	// Update the cache for the affected orders
	for _, id := range updatedIDs {
		// Construct the cache key for each order
		cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(id, 10)}

		// Attempt to retrieve the cached order
		var cachedOrder domain.PaymentOrder
		cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

		if cacheErr == nil {
			// If found in cache, update the status and save back
			cachedOrder.Status = constants.Expired
			if saveErr := c.cache.SaveItem(cacheKey, cachedOrder, c.config.GetExpiredOrderTime()); saveErr != nil {
				logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", id, saveErr)
			}
		} else {
			// Log cache miss, but proceed
			logger.GetLogger().Warnf("Payment order ID %d not found in cache: %v", id, cacheErr)
		}
	}

	return updatedIDs, nil
}

func (c *paymentOrderCache) GetPaymentOrders(
	ctx context.Context,
	limit, offset int,
	status, orderBy *string,
	orderDirection constants.OrderDirection,
) ([]domain.PaymentOrder, error) {
	// Generate the cache key based on input parameters
	cacheKey := &caching.Keyer{Raw: fmt.Sprintf("%sGetPaymentOrders_limit:%d_offset:%d_status:%v_orderBy:%v_orderDirection:%v",
		keyPrefixPaymentOrder, limit, offset, status, orderBy, orderDirection)}

	// Attempt to retrieve the data from the cache
	var paymentOrders []domain.PaymentOrder
	if err := c.cache.RetrieveItem(cacheKey, &paymentOrders); err == nil {
		// Cache hit: return the cached data
		return paymentOrders, nil
	}

	// Cache miss: fetch from the repository (DB)
	paymentOrders, err := c.paymentOrderRepository.GetPaymentOrders(
		ctx, limit, offset, status, orderBy, orderDirection,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment orders from repository: %w", err)
	}

	// Store the fetched data in the cache for future use
	if cacheErr := c.cache.SaveItem(cacheKey, paymentOrders, defaultCacheTimePaymentOrder); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache payment orders for limit %d, offset %d: %v", limit, offset, cacheErr)
	}

	return paymentOrders, nil
}

func (c *paymentOrderCache) GetPaymentOrderByID(ctx context.Context, id uint64) (*domain.PaymentOrder, error) {
	// Construct the cache key using the order ID
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(id, 10)}

	// Attempt to retrieve the payment order from the cache
	var cachedOrder domain.PaymentOrder
	if err := c.cache.RetrieveItem(cacheKey, &cachedOrder); err == nil {
		// Cache hit: return the cached order
		return &cachedOrder, nil
	} else {
		// Log cache miss
		logger.GetLogger().Infof("Cache miss for payment order ID %d: %v", id, err)
	}

	// Cache miss: fetch the payment order from the repository (DB)
	order, err := c.paymentOrderRepository.GetPaymentOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment order ID %d from repository: %w", id, err)
	}

	// Cache the fetched payment order for future use
	if cacheErr := c.cache.SaveItem(cacheKey, *order, c.config.GetExpiredOrderTime()); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache payment order ID %d: %v", id, cacheErr)
	}

	return order, nil
}
