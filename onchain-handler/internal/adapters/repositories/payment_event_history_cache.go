package repositories

import (
	"context"
	"fmt"
	"strconv"

	"github.com/genefriendway/onchain-handler/conf"
	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type paymentEventHistoryCache struct {
	paymentEventHistoryRepository repotypes.PaymentEventHistoryRepository
	cache                         cachetypes.CacheRepository
}

func NewPaymentEventHistoryCacheRepository(
	repo repotypes.PaymentEventHistoryRepository,
	cache cachetypes.CacheRepository,
) repotypes.PaymentEventHistoryRepository {
	return &paymentEventHistoryCache{
		paymentEventHistoryRepository: repo,
		cache:                         cache,
	}
}

func (c *paymentEventHistoryCache) CreatePaymentEventHistory(
	ctx context.Context,
	paymentEvents []entities.PaymentEventHistory,
) ([]entities.PaymentEventHistory, error) {
	// Create payment event history records in the repository
	createdEvents, err := c.paymentEventHistoryRepository.CreatePaymentEventHistory(ctx, paymentEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment event history in repository: %w", err)
	}

	// Group payment events by PaymentOrderID
	eventsByOrderID := make(map[uint64][]entities.PaymentEventHistory)
	for _, event := range createdEvents {
		eventsByOrderID[event.PaymentOrderID] = append(eventsByOrderID[event.PaymentOrderID], event)
	}

	// Cache each event after retrieving the associated payment order from the cache
	for orderID, events := range eventsByOrderID {
		// Construct the cache key for the payment order
		cacheKey := &cachetypes.Keyer{Raw: keyPrefixPaymentOrder + strconv.FormatUint(orderID, 10)}

		// Retrieve the associated payment order from the cache
		var cachedOrder entities.PaymentOrder
		cacheErr := c.cache.RetrieveItem(cacheKey, &cachedOrder)

		// If found in cache, process the events
		if cacheErr == nil {
			// Append the new payment event histories to the cached order
			cachedOrder.PaymentEventHistories = append(cachedOrder.PaymentEventHistories, events...)

			// Save the updated cached order back to the cache
			if err := c.cache.SaveItem(cacheKey, cachedOrder, conf.GetExpiredOrderTime()); err != nil {
				logger.GetLogger().Warnf("Failed to update cache for payment order ID %d: %v", orderID, err)
			}
		} else {
			// Log if payment order is not found in the cache
			logger.GetLogger().Warnf("Failed to retrieve payment order ID %d from cache: %v", orderID, cacheErr)
		}
	}

	// Return the created events with updated fields
	return createdEvents, nil
}
