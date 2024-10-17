package queue

import (
	"context"
	"fmt"
	"sync"
)

// Queue is a generic structure that holds items of any type.
type Queue[T comparable] struct {
	ctx     context.Context
	mu      sync.Mutex
	items   []T
	itemSet map[T]struct{} // A map to track the existence of items
	limit   int
	loader  func(ctx context.Context, limit int) ([]T, error)
}

// NewQueue creates a new generic queue with a specified limit and loader function.
func NewQueue[T comparable](ctx context.Context, limit int, loader func(ctx context.Context, limit int) ([]T, error)) (*Queue[T], error) {
	q := &Queue[T]{
		ctx:     ctx,
		items:   make([]T, 0, limit),
		itemSet: make(map[T]struct{}), // Initialize the map for tracking items
		limit:   limit,
		loader:  loader,
	}

	// Load initial items
	initialItems, err := loader(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial items: %v", err)
	}

	// Add initial items and populate the map
	for _, item := range initialItems {
		q.items = append(q.items, item)
		q.itemSet[item] = struct{}{}
	}

	return q, nil
}

// Enqueue adds a new item to the queue, ensuring that the limit is not exceeded and avoiding duplicates.
func (q *Queue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if the item already exists in the map (for faster duplicate checking)
	if _, exists := q.itemSet[item]; exists {
		return fmt.Errorf("item already exists in the queue")
	}

	if len(q.items) >= q.limit {
		return fmt.Errorf("queue is full")
	}

	// Add item to both the slice and the map
	q.items = append(q.items, item)
	q.itemSet[item] = struct{}{} // Add to map for fast lookups
	return nil
}

// Dequeue removes an item from the queue, typically when processing is complete.
func (q *Queue[T]) Dequeue(condition func(T) bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	var removed bool
	for i, item := range q.items {
		if condition(item) {
			q.items = append(q.items[:i], q.items[i+1:]...)
			delete(q.itemSet, item) // Remove from map as well
			removed = true
			break
		}
	}

	if !removed {
		return fmt.Errorf("item not found in the queue")
	}

	return nil
}

// FillQueue fills the queue up to the limit by loading more items if necessary.
func (q *Queue[T]) FillQueue() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Calculate how many more items we need to load
	itemsToLoad := q.limit - len(q.items)
	if itemsToLoad <= 0 {
		return nil // Queue is already full or at limit
	}

	// Load more items
	newItems, err := q.loader(q.ctx, itemsToLoad)
	if err != nil {
		return fmt.Errorf("failed to load more items: %w", err)
	}

	// Append new items to the queue and update the map
	for _, item := range newItems {
		if _, exists := q.itemSet[item]; !exists {
			q.items = append(q.items, item)
			q.itemSet[item] = struct{}{}
		}
	}

	return nil
}

// GetItems returns all the current items in the queue.
func (q *Queue[T]) GetItems() []T {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.items
}

// GetSmallestValue returns the smallest value based on a custom comparator function.
// The comparator should return true if the first item is smaller than the second.
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

func (q *Queue[T]) Lock() {
	q.mu.Lock()
}

func (q *Queue[T]) Unlock() {
	q.mu.Unlock()
}

func (q *Queue[T]) GetLimit() int {
	return q.limit
}
