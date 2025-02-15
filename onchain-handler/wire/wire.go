package wire

import (
	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/adapters/repositories"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/ucases"
)

// Struct to hold all repositories
type repos struct {
	BlockStateRepo           interfaces.BlockStateRepository
	PaymentOrderRepo         interfaces.PaymentOrderRepository
	PaymentWalletRepo        interfaces.PaymentWalletRepository
	PaymentWalletBalanceRepo interfaces.PaymentWalletBalanceRepository
	TokenTransferRepo        interfaces.TokenTransferRepository
	PaymentEventHistoryRepo  interfaces.PaymentEventHistoryRepository
	UserWalletRepo           interfaces.UserWalletRepository
	NetworkMetadataRepo      interfaces.NetworkMetadataRepository
	PaymentStatisticsRepo    interfaces.PaymentStatisticsRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, config *conf.Configuration, cacheRepo infrainterfaces.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		BlockStateRepo:           repositories.NewBlockstateRepository(db),
		PaymentOrderRepo:         repositories.NewPaymentOrderCacheRepository(repositories.NewPaymentOrderRepository(db, config), cacheRepo, config),
		PaymentWalletRepo:        repositories.NewPaymentWalletRepository(db, config),
		PaymentWalletBalanceRepo: repositories.NewPaymentWalletBalanceRepository(db),
		TokenTransferRepo:        repositories.NewTokenTransferRepository(db),
		PaymentEventHistoryRepo:  repositories.NewPaymentEventHistoryCacheRepository(repositories.NewPaymentEventHistoryRepository(db), cacheRepo, config),
		UserWalletRepo:           repositories.NewUserWalletRepository(db),
		NetworkMetadataRepo:      repositories.NewNetworkMetadataCacheRepository(repositories.NewNetworkMetadataRepository(db), cacheRepo),
		PaymentStatisticsRepo:    repositories.NewPaymentStatisticsRepository(db),
	}
}

// Struct to hold all use cases
type UseCases struct {
	BlockStateUCase          interfaces.BlockStateUCase
	PaymentOrderUCase        interfaces.PaymentOrderUCase
	TokenTransferUCase       interfaces.TokenTransferUCase
	PaymentEventHistoryUCase interfaces.PaymentEventHistoryUCase
	PaymentWalletUCase       interfaces.PaymentWalletUCase
	UserWalletUCase          interfaces.UserWalletUCase
	NetworkMetadataUCase     interfaces.NetworkMetadataUCase
	PaymentStatisticsUCase   interfaces.PaymentStatisticsUCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, config *conf.Configuration, cacheRepo infrainterfaces.CacheRepository) *UseCases {
	repos := initializeRepos(db, config, cacheRepo)

	// Return all use cases
	return &UseCases{
		BlockStateUCase:          ucases.NewBlockStateUCase(repos.BlockStateRepo),
		PaymentOrderUCase:        ucases.NewPaymentOrderUCase(db, repos.PaymentOrderRepo, repos.PaymentWalletRepo, repos.BlockStateRepo, repos.PaymentStatisticsRepo, cacheRepo, config),
		TokenTransferUCase:       ucases.NewTokenTransferUCase(repos.TokenTransferRepo, config),
		PaymentEventHistoryUCase: ucases.NewPaymentEventHistoryUCase(repos.PaymentEventHistoryRepo),
		PaymentWalletUCase:       ucases.NewPaymentWalletUCase(db, repos.PaymentWalletRepo, repos.PaymentWalletBalanceRepo, config),
		UserWalletUCase:          ucases.NewUserWalletUCase(repos.UserWalletRepo),
		NetworkMetadataUCase:     ucases.NewNetworkMetadataUCase(repos.NetworkMetadataRepo),
		PaymentStatisticsUCase:   ucases.NewPaymentStatisticsCase(repos.PaymentStatisticsRepo),
	}
}
