package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

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

func (c *paymentOrderCache) CreatePaymentOrders(
	tx *gorm.DB,
	ctx context.Context,
	orders []domain.PaymentOrder,
) ([]domain.PaymentOrder, error) {
	// Validate input
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders to create")
	}

	// Create orders in the repository (DB)
	createdOrders, err := c.paymentOrderRepository.CreatePaymentOrders(tx, ctx, orders)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment orders in repository: %w", err)
	}

	// Cache the created orders
	var cacheErrors []error
	for _, order := range createdOrders {
		cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(order.ID, 10)}
		if err := c.cache.SaveItem(cacheKey, order, c.config.GetExpiredOrderTime()); err != nil {
			logger.GetLogger().Warnf("Failed to cache payment order ID %d: %v", order.ID, err)
			cacheErrors = append(cacheErrors, err)
		}
	}

	// Log and return any cache errors
	if len(cacheErrors) > 0 {
		logger.GetLogger().Warnf("Encountered %d cache errors while caching payment orders", len(cacheErrors))
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
	if src.BlockHeight != 0 {
		dst.BlockHeight = src.BlockHeight
	}
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

func (c *paymentOrderCache) BatchUpdateOrdersToExpired(
	ctx context.Context,
	orderIDs []uint64,
) error {
	// Iterate over the order IDs and new statuses
	for _, orderID := range orderIDs {
		// Construct the cache key
		cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

		// Attempt to retrieve the order from the cache
		var cachedOrder domain.PaymentOrder
		cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

		// If found in cache, update it
		if cacheErr == nil {
			// Update status in the cached order
			cachedOrder.Status = constants.Expired

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
	if err := c.paymentOrderRepository.BatchUpdateOrdersToExpired(ctx, orderIDs); err != nil {
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
	paymentOrders, err := c.paymentOrderRepository.GetExpiredPaymentOrders(ctx, network, c.config.GetOrderCutoffTime())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expired payment orders: %w", err)
	}

	// Store the fetched data in the cache for future use
	if cacheErr := c.cache.SaveItem(key, paymentOrders, defaultCacheTimePaymentOrder); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache expired payment orders for network %s: %v", network, cacheErr)
	}

	return paymentOrders, nil
}

func (c *paymentOrderCache) UpdateExpiredOrdersToFailed(ctx context.Context) ([]uint64, error) {
	// Call the repository to update expired orders to "Failed" and get the updated IDs
	updatedIDs, err := c.paymentOrderRepository.UpdateExpiredOrdersToFailed(ctx, c.config.GetOrderCutoffTime())
	if err != nil {
		return nil, fmt.Errorf("failed to update expired orders to failed in repository: %w", err)
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
			cachedOrder.Status = constants.Failed
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

func (c *paymentOrderCache) UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error) {
	// Call repository to update active orders to expired and get the updated IDs
	updatedIDs, err := c.paymentOrderRepository.UpdateActiveOrdersToExpired(ctx, c.config.GetOrderCutoffTime())
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
	requestIDs []string,
	status, orderBy *string,
	orderDirection constants.OrderDirection,
) ([]domain.PaymentOrder, error) {
	// Generate the cache key based on input parameters
	requestIDsKey := strings.Join(requestIDs, ",")
	cacheKey := &caching.Keyer{
		Raw: fmt.Sprintf("%sGetPaymentOrders_limit:%d_offset:%d_requestIDs:%s_status:%v_orderBy:%v_orderDirection:%v",
			keyPrefixPaymentOrder, limit, offset, requestIDsKey, status, orderBy, orderDirection),
	}

	// Attempt to retrieve the data from the cache
	var paymentOrders []domain.PaymentOrder
	if err := c.cache.RetrieveItem(cacheKey, &paymentOrders); err == nil {
		// Cache hit: return the cached data
		return paymentOrders, nil
	}

	// Cache miss: fetch from the repository (DB)
	paymentOrders, err := c.paymentOrderRepository.GetPaymentOrders(
		ctx, limit, offset, requestIDs, status, orderBy, orderDirection,
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

func (c *paymentOrderCache) GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]domain.PaymentOrder, error) {
	// Construct the cache key using the order IDs
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + fmt.Sprint(ids)}

	// Attempt to retrieve the payment orders from the cache
	var cachedOrders []domain.PaymentOrder
	if err := c.cache.RetrieveItem(cacheKey, &cachedOrders); err == nil {
		// Cache hit: return the cached orders
		return cachedOrders, nil
	} else {
		// Log cache miss
		logger.GetLogger().Infof("Cache miss for payment orders IDs %v: %v", ids, err)
	}

	// Cache miss: fetch the payment orders from the repository (DB)
	orders, err := c.paymentOrderRepository.GetPaymentOrdersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment orders with IDs %v from repository: %w", ids, err)
	}

	// Cache the fetched payment orders for future use
	if cacheErr := c.cache.SaveItem(cacheKey, orders, c.config.GetExpiredOrderTime()); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache payment orders IDs %v: %v", ids, cacheErr)
	}

	return orders, nil
}

func (c *paymentOrderCache) GetPaymentOrderByRequestID(ctx context.Context, requestID string) (*domain.PaymentOrder, error) {
	// Fetch the order ID using the request ID
	orderID, err := c.GetPaymentOrderIDByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Construct the cache key using the order ID
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

	// Attempt to retrieve the payment order from the cache
	var cachedOrder domain.PaymentOrder
	if err := c.cache.RetrieveItem(cacheKey, &cachedOrder); err == nil {
		// Cache hit: return the cached order
		return &cachedOrder, nil
	} else {
		// Log cache miss
		logger.GetLogger().Infof("Cache miss for payment order Request ID %d: %v", requestID, err)
	}

	// Cache miss: fetch the payment order from the repository (DB)
	order, err := c.paymentOrderRepository.GetPaymentOrderByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment order with Request ID %s from repository: %w", requestID, err)
	}

	// Cache the fetched payment order for future use
	if cacheErr := c.cache.SaveItem(cacheKey, *order, c.config.GetExpiredOrderTime()); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache payment order with Request ID %d: %v", requestID, cacheErr)
	}

	return order, nil
}

func (c *paymentOrderCache) GetPaymentOrderIDByRequestID(ctx context.Context, requestID string) (uint64, error) {
	// Construct the cache key using the request ID
	cacheKey := &caching.Keyer{Raw: keyPrefixPaymentOrder + requestID}

	// Attempt to retrieve the payment order ID from the cache
	var cachedOrderID uint64
	if err := c.cache.RetrieveItem(cacheKey, &cachedOrderID); err == nil {
		// Cache hit: return the cached order ID
		return cachedOrderID, nil
	} else {
		// Log cache miss
		logger.GetLogger().Infof("Cache miss for payment order ID with Request ID %s: %v", requestID, err)
	}

	// Cache miss: fetch the payment order ID from the repository (DB)
	orderID, err := c.paymentOrderRepository.GetPaymentOrderIDByRequestID(ctx, requestID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch payment order ID with Request ID %s from repository: %w", requestID, err)
	}

	// Cache the fetched payment order ID for future use
	if cacheErr := c.cache.SaveItem(cacheKey, orderID, c.config.GetExpiredOrderTime()); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to cache payment order ID with Request ID %s: %v", requestID, cacheErr)
	}

	return orderID, nil
}
