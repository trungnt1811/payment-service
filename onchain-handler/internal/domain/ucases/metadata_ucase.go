package ucases

import (
	"context"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

type metadataUCase struct {
	networkMetadataRepository repotypes.NetworkMetadataRepository
	tokenMetadataRepository   repotypes.TokenMetadataRepository
}

func NewMetadataUCase(
	networkMetadataRepository repotypes.NetworkMetadataRepository,
	tokenMetadataRepository repotypes.TokenMetadataRepository,
) ucasetypes.MetadataUCase {
	return &metadataUCase{
		networkMetadataRepository: networkMetadataRepository,
		tokenMetadataRepository:   tokenMetadataRepository,
	}
}

func (u *metadataUCase) GetNetworksMetadata(ctx context.Context) ([]dto.NetworkMetadataDTO, error) {
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

func (u *metadataUCase) GetTokensMetadata(ctx context.Context) ([]dto.TokenMetadataDTO, error) {
	tokensMetadata, err := u.tokenMetadataRepository.GetTokensMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var tokensMetadataDTO []dto.TokenMetadataDTO
	for _, tokenMetadata := range tokensMetadata {
		tokensMetadataDTO = append(tokensMetadataDTO, tokenMetadata.ToDto())
	}

	return tokensMetadataDTO, nil
}
