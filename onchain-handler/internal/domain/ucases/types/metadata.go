package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type MetadataUCase interface {
	GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error)
	GetTokensMetadata(ctx context.Context) ([]dto.TokenMetadataDTO, error)
}
