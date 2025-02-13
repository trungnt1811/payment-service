package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type NetworkMetadataRepository interface {
	GetNetworksMetadata(ctx context.Context) ([]entities.NetworkMetadata, error)
}

type NetworkMetadataUCase interface {
	GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error)
}
