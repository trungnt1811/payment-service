package set

import (
	"context"
	"fmt"
	"sync"

	"github.com/genefriendway/onchain-handler/infra/set/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type set[T comparable] struct {
	ctx     context.Context
	mu      sync.Mutex
	items   map[string]T
	keyFunc func(T) string
}

// NewSet creates a new set.
func NewSet[T comparable](ctx context.Context, keyFunc func(T) string) (types.Set[T], error) {
	return &set[T]{
		ctx:     ctx,
		items:   make(map[string]T),
		keyFunc: keyFunc,
	}, nil
}

// GetAll returns a copy of all items.
func (s *set[T]) GetAll() []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]T, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

// Contains checks if an item exists in the set.
func (s *set[T]) Contains(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.items[key]
	return exists
}

// GetItem retrieves an item by its key.
func (s *set[T]) GetItem(key string) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, exists := s.items[key]
	return item, exists
}

// Add inserts an item into the set.
func (s *set[T]) Add(item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.keyFunc(item)
	if _, exists := s.items[key]; exists {
		return fmt.Errorf("item %v already exists", item)
	}

	s.items[key] = item
	return nil
}

// Remove deletes an item from the set based on a condition.
func (s *set[T]) Remove(condition func(T) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.items {
		if condition(item) {
			delete(s.items, key)
			return true // Successfully removed
		}
	}
	return false // No matching item found
}

// UpdateItem replaces an existing item in the set.
func (s *set[T]) UpdateItem(key string, newItem T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[key]; !exists {
		return fmt.Errorf("item with key %s not found", key)
	}

	s.items[key] = newItem
	return nil
}

// Fill loads additional items into the set using a provided loader function.
func (s *set[T]) Fill(loader func(ctx context.Context) ([]T, error)) error {
	newItems, err := loader(s.ctx) // Use passed loader instead of struct's loader
	if err != nil {
		logger.GetLogger().Errorf("Error loading items: %v", err)
		return fmt.Errorf("failed to load more items: %w", err)
	}

	logger.GetLogger().Debugf("Loaded %d new items", len(newItems))

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range newItems {
		s.items[s.keyFunc(item)] = item
	}

	return nil
}
