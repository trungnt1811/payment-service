package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type TokenMetadataRepository interface {
	GetTokensMetadata(ctx context.Context) ([]entities.TokenMetadata, error)
}
