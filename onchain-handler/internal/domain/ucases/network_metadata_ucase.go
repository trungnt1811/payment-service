package ucases

import (
	"context"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

type networkMetadataUCase struct {
	networkMetadataRepository repotypes.NetworkMetadataRepository
}

func NewNetworkMetadataUCase(
	networkMetadataRepository repotypes.NetworkMetadataRepository,
) ucasetypes.NetworkMetadataUCase {
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
