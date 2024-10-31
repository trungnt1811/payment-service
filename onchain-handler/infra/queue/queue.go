package queue

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/log"
)

type Queue[T comparable] struct {
	ctx     context.Context
	mu      sync.Mutex
	items   []T
	itemSet map[T]struct{}
	limit   int
	loader  func(ctx context.Context, limit, offset int) ([]T, error)
}

// NewQueue creates a new queue with an initial load of items.
func NewQueue[T comparable](ctx context.Context, limit int, loader func(ctx context.Context, limit, offset int) ([]T, error)) (*Queue[T], error) {
	q := &Queue[T]{
		ctx:     ctx,
		items:   make([]T, 0, limit),
		itemSet: make(map[T]struct{}),
		limit:   limit,
		loader:  loader,
	}

	if err := q.loadInitialItems(); err != nil {
		return nil, err
	}
	return q, nil
}

// loadInitialItems populates the queue with the first set of items.
func (q *Queue[T]) loadInitialItems() error {
	items, err := q.loader(q.ctx, q.limit, 0)
	if err != nil {
		return fmt.Errorf("failed to load initial items: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range items {
		q.items = append(q.items, item)
		q.itemSet[item] = struct{}{}
	}
	return nil
}

// Enqueue adds a new item, ensuring no duplicates and enforcing the limit.
func (q *Queue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.itemSet[item]; exists {
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
	q.itemSet[item] = struct{}{}
	return nil
}

// isNilable checks if a value is of a nilable type.
func isNilable(item interface{}) bool {
	// Get the kind of the item
	kind := reflect.TypeOf(item).Kind()
	return kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Map || kind == reflect.Func || kind == reflect.Chan || kind == reflect.Interface
}

// Dequeue removes the first item that matches the condition.
func (q *Queue[T]) Dequeue(condition func(T) bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if condition(item) {
			log.LG.Infof("Dequeueing item: %v", item)
			// Fast removal, maintaining order
			copy(q.items[i:], q.items[i+1:])
			q.items = q.items[:len(q.items)-1]
			delete(q.itemSet, item)

			// Check if shrinking is necessary
			if q.limit > constants.MinQueueLimit {
				q.maybeShrink()
			}
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

// FillQueue loads more items to fill up to the limit.
func (q *Queue[T]) FillQueue() error {
	q.mu.Lock()

	// If current items exceed the MaxQueueLimit, log a message and return early
	if len(q.items) >= constants.MaxQueueLimit {
		log.LG.Infof("Queue size has reached the maximum limit of %d items. No additional items will be loaded.", constants.MaxQueueLimit)
		q.mu.Unlock()
		return nil
	}

	offset := len(q.items)
	remainingCapacity := q.limit - offset
	q.mu.Unlock()

	// Load more items
	newItems, err := q.loader(q.ctx, remainingCapacity+1, offset)
	if err != nil {
		log.LG.Infof("Error loading items: %v", err)
		return fmt.Errorf("failed to load more items: %w", err)
	}
	log.LG.Infof("Loaded %d new items", len(newItems))

	q.mu.Lock()
	defer q.mu.Unlock()

	// Scale if necessary
	if len(newItems) > remainingCapacity {
		q.scaleQueue()
	}

	// Add unique new items
	for _, item := range newItems {
		if _, exists := q.itemSet[item]; !exists {
			q.items = append(q.items, item)
			q.itemSet[item] = struct{}{}
		}
	}
	return nil
}

func (q *Queue[T]) scaleQueue() {
	if q.limit >= constants.MaxQueueLimit {
		q.limit = constants.MaxQueueLimit
		log.LG.Infof("Reached max queue limit: %d", q.limit)
		return
	}

	// Adjust new limit based on scale factor
	newLimit := int(float64(q.limit)*constants.ScaleFactor + 1)
	if newLimit > constants.MaxQueueLimit {
		newLimit = constants.MaxQueueLimit
	}
	q.limit = newLimit
	log.LG.Infof("Scaling queue to new limit: %d", q.limit)
}

// maybeShrink reduces the queue's size if it's consistently underutilized.
func (q *Queue[T]) maybeShrink() {
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
	log.LG.Infof("Shrinking queue to new limit: %d", q.limit)
}

// GetItems returns a copy of all current items.
func (q *Queue[T]) GetItems() []T {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Create a copy of the items
	itemsCopy := make([]T, len(q.items))
	copy(itemsCopy, q.items)
	return itemsCopy
}

// ReplaceItemAtIndex replaces an item at a specific index with a new item.
func (q *Queue[T]) ReplaceItemAtIndex(index int, newItem T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check for index bounds
	if index < 0 || index >= len(q.items) {
		return fmt.Errorf("index out of bounds")
	}

	// Replace the item
	oldItem := q.items[index]
	q.items[index] = newItem

	// Update itemSet
	delete(q.itemSet, oldItem)
	q.itemSet[newItem] = struct{}{}

	return nil
}

// GetSmallestValue finds the smallest value according to a comparator function.
func (q *Queue[T]) GetSmallestValue(compare func(T, T) bool) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		var zero T
		return zero, fmt.Errorf("queue is empty")
	}

	smallest := q.items[0]
	for _, item := range q.items[1:] {
		if compare(item, smallest) {
			smallest = item
		}
	}

	return smallest, nil
}

func (q *Queue[T]) GetLimit() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.limit
}
