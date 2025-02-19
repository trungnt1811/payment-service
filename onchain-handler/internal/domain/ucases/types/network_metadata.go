package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type NetworkMetadataUCase interface {
	GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error)
}
