package types

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// EventHandler is a type for event handler functions.
type EventHandler func(log types.Log) (any, error)

type BaseEventListener interface {
	RunListener(ctx context.Context) error
	RegisterConfirmedEventListener(contractAddress string, handler EventHandler)
	RegisterRealtimeEventListener(contractAddress string, handler EventHandler)
}

type EventListener interface {
	Register(ctx context.Context)
}
