package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type NetworkMetadataRepository interface {
	GetNetworksMetadata(ctx context.Context) ([]entities.NetworkMetadata, error)
}
