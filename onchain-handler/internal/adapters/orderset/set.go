package orderset

import (
	"context"
	"errors"
	"fmt"
	"time"

	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	"github.com/genefriendway/onchain-handler/internal/adapters/orderset/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type set[T comparable] struct {
	ctx       context.Context
	cacheRepo cachetypes.CacheRepository
	keyFunc   func(T) string
	ttl       time.Duration
}

// NewSet creates a new set.
func NewSet[T comparable](
	ctx context.Context,
	keyFunc func(T) string,
	cacheRepo cachetypes.CacheRepository,
) (types.Set[T], error) {
	return &set[T]{
		ctx:       ctx,
		cacheRepo: cacheRepo,
		keyFunc:   keyFunc,
		ttl:       -1, // Default no expiration
	}, nil
}

func (s *set[T]) prefixKeyOnly() fmt.Stringer {
	return &cachetypes.Keyer{Raw: "orderset_"}
}

func (s *set[T]) fullKey(key string) string {
	return fmt.Sprintf("orderset_%s", key)
}

// GetAll returns a copy of all items.
func (s *set[T]) GetAll() []T {
	valFactory := func() any {
		var zero T
		return &zero
	}

	itemsAny, err := s.cacheRepo.GetAllMatching(s.prefixKeyOnly(), valFactory)
	if err != nil {
		logger.GetLogger().Errorf("GetAll: failed to fetch all matching keys: %v", err)
		return nil
	}

	results := make([]T, 0, len(itemsAny))
	for _, item := range itemsAny {
		if v, ok := item.(*T); ok {
			results = append(results, *v)
		} else {
			logger.GetLogger().Warnf("GetAll: failed to cast item %v to type %T", item, *new(T))
		}
	}
	return results
}

// Contains checks if an item exists in the set.
func (s *set[T]) Contains(key string) bool {
	var dummy T
	err := s.cacheRepo.RetrieveItem(&cachetypes.Keyer{Raw: s.fullKey(key)}, &dummy)
	return err == nil
}

// GetItem retrieves an item by its key.
func (s *set[T]) GetItem(key string) (T, bool) {
	var item T
	err := s.cacheRepo.RetrieveItem(&cachetypes.Keyer{Raw: s.fullKey(key)}, &item)
	if err != nil {
		return item, false
	}
	return item, true
}

// Add inserts an item into the set.
func (s *set[T]) Add(item T) error {
	key := s.fullKey(s.keyFunc(item))

	var existing T
	err := s.cacheRepo.RetrieveItem(&cachetypes.Keyer{Raw: key}, &existing)
	if err == nil {
		return fmt.Errorf("item %v already exists", item)
	}
	if !errors.Is(err, cachetypes.ErrNotFound) {
		return fmt.Errorf("failed to check existing item: %w", err)
	}

	return s.cacheRepo.SaveItem(&cachetypes.Keyer{Raw: key}, item, s.ttl)
}

// Remove deletes an item from the set based on a condition.
func (s *set[T]) Remove(condition func(T) bool) bool {
	items, err := s.cacheRepo.GetAllMatching(s.prefixKeyOnly(), func() any {
		var t T
		return &t
	})
	if err != nil {
		logger.GetLogger().Errorf("Remove: failed to get all items: %v", err)
		return false
	}

	for _, raw := range items {
		item := *(raw.(*T)) // convert back from `any` to `T`
		if condition(item) {
			key := s.fullKey(s.keyFunc(item))
			if err := s.cacheRepo.RemoveItem(&cachetypes.Keyer{Raw: key}); err != nil {
				logger.GetLogger().Errorf("Remove: failed to delete item %v: %v", item, err)
				continue
			}
			return true
		}
	}
	return false
}

// UpdateItem replaces an existing item in the set.
func (s *set[T]) UpdateItem(key string, newItem T) error {
	fullKey := s.fullKey(key)

	var existing T
	err := s.cacheRepo.RetrieveItem(&cachetypes.Keyer{Raw: fullKey}, &existing)
	if err != nil {
		if errors.Is(err, cachetypes.ErrNotFound) {
			return fmt.Errorf("item with key %s not found", key)
		}
		return fmt.Errorf("failed to retrieve existing item: %w", err)
	}

	if err := s.cacheRepo.SaveItem(&cachetypes.Keyer{Raw: fullKey}, newItem, s.ttl); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

// Fill loads additional items into the set using a provided loader function.
func (s *set[T]) Fill(loader func(ctx context.Context) ([]T, error)) error {
	newItems, err := loader(s.ctx)
	if err != nil {
		logger.GetLogger().Errorf("Error loading items: %v", err)
		return fmt.Errorf("failed to load more items: %w", err)
	}

	logger.GetLogger().Debugf("Loaded %d new items", len(newItems))

	for _, item := range newItems {
		key := s.fullKey(s.keyFunc(item))

		var existing T
		err := s.cacheRepo.RetrieveItem(&cachetypes.Keyer{Raw: key}, &existing)
		if err == nil {
			// Item already exists → skip
			continue
		}

		if !errors.Is(err, cachetypes.ErrNotFound) {
			logger.GetLogger().Warnf("Fill: failed to check existence for item %v: %v", item, err)
			continue
		}

		if err := s.cacheRepo.SaveItem(&cachetypes.Keyer{Raw: key}, item, s.ttl); err != nil {
			logger.GetLogger().Warnf("Fill: failed to save item %v: %v", item, err)
		}
	}

	return nil
}
