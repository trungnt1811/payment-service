package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

type BaseEventListener interface {
	RunListener(ctx context.Context) error
	RegisterEventListener(contractAddress string, handler func(log types.Log) (interface{}, error))
}

type EventListener interface {
	Register(ctx context.Context)
}
