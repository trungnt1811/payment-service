package wire

import (
	"gorm.io/gorm"

	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	"github.com/genefriendway/onchain-handler/infra/set/types"
	"github.com/genefriendway/onchain-handler/internal/adapters/repositories"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/ucases"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

// Struct to hold all repositories
type repos struct {
	BlockStateRepo           repotypes.BlockStateRepository
	PaymentOrderRepo         repotypes.PaymentOrderRepository
	PaymentWalletRepo        repotypes.PaymentWalletRepository
	PaymentWalletBalanceRepo repotypes.PaymentWalletBalanceRepository
	TokenTransferRepo        repotypes.TokenTransferRepository
	PaymentEventHistoryRepo  repotypes.PaymentEventHistoryRepository
	NetworkMetadataRepo      repotypes.NetworkMetadataRepository
	PaymentStatisticsRepo    repotypes.PaymentStatisticsRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, cacheRepo cachetypes.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		BlockStateRepo:           repositories.NewBlockstateRepository(db),
		PaymentOrderRepo:         repositories.NewPaymentOrderCacheRepository(repositories.NewPaymentOrderRepository(db), cacheRepo),
		PaymentWalletRepo:        repositories.NewPaymentWalletRepository(db),
		PaymentWalletBalanceRepo: repositories.NewPaymentWalletBalanceRepository(db),
		TokenTransferRepo:        repositories.NewTokenTransferRepository(db),
		PaymentEventHistoryRepo:  repositories.NewPaymentEventHistoryCacheRepository(repositories.NewPaymentEventHistoryRepository(db), cacheRepo),
		NetworkMetadataRepo:      repositories.NewNetworkMetadataCacheRepository(repositories.NewNetworkMetadataRepository(db), cacheRepo),
		PaymentStatisticsRepo:    repositories.NewPaymentStatisticsRepository(db),
	}
}

// Struct to hold all use cases
type UseCases struct {
	BlockStateUCase          ucasetypes.BlockStateUCase
	PaymentOrderUCase        ucasetypes.PaymentOrderUCase
	TokenTransferUCase       ucasetypes.TokenTransferUCase
	PaymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase
	PaymentWalletUCase       ucasetypes.PaymentWalletUCase
	NetworkMetadataUCase     ucasetypes.NetworkMetadataUCase
	PaymentStatisticsUCase   ucasetypes.PaymentStatisticsUCase
}

// Initialize use cases
func InitializeUseCases(
	db *gorm.DB, cacheRepo cachetypes.CacheRepository, paymentOrderSet types.Set[dto.PaymentOrderDTO],
) *UseCases {
	repos := initializeRepos(db, cacheRepo)

	// Return all use cases
	return &UseCases{
		BlockStateUCase: ucases.NewBlockStateUCase(repos.BlockStateRepo),
		PaymentOrderUCase: ucases.NewPaymentOrderUCase(
			db,
			repos.PaymentOrderRepo,
			repos.PaymentWalletRepo,
			repos.BlockStateRepo,
			repos.PaymentStatisticsRepo,
			cacheRepo,
			paymentOrderSet,
		),
		TokenTransferUCase:       ucases.NewTokenTransferUCase(repos.TokenTransferRepo),
		PaymentEventHistoryUCase: ucases.NewPaymentEventHistoryUCase(repos.PaymentEventHistoryRepo),
		PaymentWalletUCase:       ucases.NewPaymentWalletUCase(db, repos.PaymentWalletRepo, repos.PaymentWalletBalanceRepo),
		NetworkMetadataUCase:     ucases.NewNetworkMetadataUCase(repos.NetworkMetadataRepo),
		PaymentStatisticsUCase:   ucases.NewPaymentStatisticsCase(repos.PaymentStatisticsRepo),
	}
}
