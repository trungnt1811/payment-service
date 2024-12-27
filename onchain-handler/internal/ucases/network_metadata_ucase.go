package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type networkMetadataUCase struct {
	networkMetadataRepository interfaces.NetworkMetadataRepository
}

func NewNetworkMetadataUCase(
	networkMetadataRepository interfaces.NetworkMetadataRepository,
) interfaces.NetworkMetadataUCase {
	return &networkMetadataUCase{
		networkMetadataRepository: networkMetadataRepository,
	}
}

func (u *networkMetadataUCase) GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error) {
	networksMetadata, err := u.networkMetadataRepository.GetNetworksMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var networksMetadataDTO []dto.NetworkMetadataDTO
	for _, networkMetadata := range networksMetadata {
		networksMetadataDTO = append(networksMetadataDTO, networkMetadata.ToDto())
	}

	return networksMetadataDTO, nil
}
