package queue

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type queue[T comparable] struct {
	ctx         context.Context
	mu          sync.Mutex
	items       []T
	keyIndexMap map[string]int
	limit       int
	loader      func(ctx context.Context, limit, offset int) ([]T, error)
	keyFunc     func(T) string
}

// NewQueue creates a new queue with an initial load of items.
func NewQueue[T comparable](
	ctx context.Context, limit int, keyFunc func(T) string, loader func(ctx context.Context, limit, offset int) ([]T, error),
) (interfaces.Queue[T], error) {
	q := queue[T]{
		ctx:         ctx,
		items:       make([]T, 0, limit),
		keyIndexMap: make(map[string]int),
		limit:       limit,
		loader:      loader,
		keyFunc:     keyFunc,
	}

	if err := q.loadInitialItems(); err != nil {
		return nil, err
	}
	return &q, nil
}

// loadInitialItems populates the queue with the first set of items.
func (q *queue[T]) loadInitialItems() error {
	items, err := q.loader(q.ctx, q.limit, 0)
	if err != nil {
		return fmt.Errorf("failed to load initial items: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for index, item := range items {
		key := q.keyFunc(item)
		q.items = append(q.items, item)
		q.keyIndexMap[key] = index
	}
	return nil
}

// Enqueue adds a new item, ensuring no duplicates and enforcing the limit.
func (q *queue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := q.keyFunc(item)
	if _, exists := q.keyIndexMap[key]; exists {
		return fmt.Errorf("item %v already exists in the queue", item)
	}

	if len(q.items) >= q.limit {
		return fmt.Errorf("queue is full")
	}

	// Check if the item is nil (only for types that can be nil)
	if isNilable(item) && reflect.ValueOf(item).IsNil() {
		return fmt.Errorf("cannot enqueue a nil item")
	}

	q.items = append(q.items, item)
	q.keyIndexMap[key] = len(q.items) - 1
	return nil
}

// isNilable checks if a value is of a nilable type.
func isNilable(item interface{}) bool {
	// Get the kind of the item
	kind := reflect.TypeOf(item).Kind()
	return kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Map || kind == reflect.Func || kind == reflect.Chan || kind == reflect.Interface
}

// Dequeue removes the first item that matches the condition.
func (q *queue[T]) Dequeue(condition func(T) bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if condition(item) {
			logger.GetLogger().Infof("Dequeueing item: %v", item)

			// Fast removal, maintaining order
			copy(q.items[i:], q.items[i+1:])
			q.items = q.items[:len(q.items)-1]

			// Remove from map
			key := q.keyFunc(item)
			delete(q.keyIndexMap, key)

			// Update indexes
			for j := i; j < len(q.items); j++ {
				q.keyIndexMap[q.keyFunc(q.items[j])] = j
			}

			// Check if shrinking is necessary
			if q.limit > constants.MinQueueLimit {
				q.maybeShrink()
			}
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

// GetIndex retrieves the index of an item by its key.
func (q *queue[T]) GetIndex(key string) (int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	index, exists := q.keyIndexMap[key]
	return index, exists
}

// FillQueue loads more items to fill up to the limit.
func (q *queue[T]) FillQueue() error {
	q.mu.Lock()

	// If current items exceed the MaxQueueLimit, log a message and return early
	if len(q.items) >= constants.MaxQueueLimit {
		logger.GetLogger().Infof("Queue size has reached the maximum limit of %d items. No additional items will be loaded.", constants.MaxQueueLimit)
		q.mu.Unlock()
		return nil
	}

	offset := len(q.items)
	remainingCapacity := q.limit - offset
	q.mu.Unlock()

	// Load more items
	newItems, err := q.loader(q.ctx, remainingCapacity+1, offset)
	if err != nil {
		logger.GetLogger().Infof("Error loading items: %v", err)
		return fmt.Errorf("failed to load more items: %w", err)
	}
	logger.GetLogger().Infof("Loaded %d new items", len(newItems))

	q.mu.Lock()
	defer q.mu.Unlock()

	// Scale if necessary
	if len(newItems) > remainingCapacity {
		q.scaleQueue()
	}

	// Add unique new items
	for _, item := range newItems {
		key := q.keyFunc(item)
		if _, exists := q.keyIndexMap[key]; !exists {
			q.items = append(q.items, item)
			q.keyIndexMap[key] = len(q.items) - 1
		}
	}
	return nil
}

func (q *queue[T]) scaleQueue() {
	if q.limit >= constants.MaxQueueLimit {
		q.limit = constants.MaxQueueLimit
		logger.GetLogger().Infof("Reached max queue limit: %d", q.limit)
		return
	}

	// +1 to avoid infinite scaling when the limit is 1
	newLimit := int(float64(q.limit)*constants.ScaleFactor + 1)
	if newLimit > constants.MaxQueueLimit {
		newLimit = constants.MaxQueueLimit
	}
	q.limit = newLimit
	logger.GetLogger().Infof("Scaling queue to new limit: %d", q.limit)
}

// maybeShrink reduces the queue's size if it's consistently underutilized.
func (q *queue[T]) maybeShrink() {
	if len(q.items) >= int(float64(q.limit)*constants.ShrinkThreshold) {
		return
	}

	if q.limit > constants.MinQueueLimit {
		newLimit := int(float64(q.limit) * constants.ShrinkFactor)
		if newLimit < constants.MinQueueLimit {
			q.limit = constants.MinQueueLimit
		} else {
			q.limit = newLimit
		}
	}
	logger.GetLogger().Infof("Shrinking queue to new limit: %d", q.limit)
}

// GetItems returns a copy of all current items.
func (q *queue[T]) GetItems() []T {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Create a copy of the items
	itemsCopy := make([]T, len(q.items))
	copy(itemsCopy, q.items)
	return itemsCopy
}

// ReplaceItemAtIndex replaces an item at a specific index with a new item.
func (q *queue[T]) ReplaceItemAtIndex(index int, newItem T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check for index bounds
	if index < 0 || index >= len(q.items) {
		return fmt.Errorf("index out of bounds")
	}

	// Replace the item
	oldItem := q.items[index]
	oldKey := q.keyFunc(oldItem)
	newKey := q.keyFunc(newItem)

	q.items[index] = newItem

	// Update itemSet
	delete(q.keyIndexMap, oldKey)
	q.keyIndexMap[newKey] = index

	return nil
}

// GetItemAtIndex retrieves a copy of an item at a specific index.
func (q *queue[T]) GetItemAtIndex(index int) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check for index bounds
	if index < 0 || index >= len(q.items) {
		var zeroValue T // Return zero-value of type T
		return zeroValue, fmt.Errorf("index out of bounds")
	}

	// Return a copy of the item to avoid external modifications
	itemCopy := q.items[index]
	return itemCopy, nil
}
