package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type NetworkMetadataRepository interface {
	GetNetworksMetadata(ctx context.Context) ([]domain.NetworkMetadata, error)
}

type NetworkMetadataUCase interface {
	GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error)
}
