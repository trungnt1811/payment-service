//go:build wireinject

package wire

import (
	"github.com/genefriendway/onchain-handler/conf"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/adapters/repositories"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/ucases"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var blockStateUCaseSet = wire.NewSet(
	repositories.NewBlockstateRepository,
	ucases.NewBlockStateUCase,
)

var paymentOrderUCaseSet = wire.NewSet(
	repositories.NewPaymentOrderRepository,
	repositories.NewPaymentOrderCacheRepository,
	repositories.NewPaymentWalletRepository,
	repositories.NewBlockstateRepository,
	repositories.NewPaymentStatisticsRepository,
	ucases.NewPaymentOrderUCase,
)

var tokenTransferUCaseSet = wire.NewSet(
	repositories.NewTokenTransferRepository,
	ucases.NewTokenTransferUCase,
)

var paymentEventHistoryUCaseSet = wire.NewSet(
	repositories.NewPaymentEventHistoryRepository,
	repositories.NewPaymentEventHistoryCacheRepository,
	ucases.NewPaymentEventHistoryUCase,
)

var paymentWalletUCaseSet = wire.NewSet(
	repositories.NewPaymentWalletRepository,
	repositories.NewPaymentWalletBalanceRepository,
	ucases.NewPaymentWalletUCase,
)

var userWalletUCaseSet = wire.NewSet(
	repositories.NewUserWalletRepository,
	ucases.NewUserWalletUCase,
)

var networkMetadataUCaseSet = wire.NewSet(
	repositories.NewNetworkMetadataRepository,
	repositories.NewNetworkMetadataCacheRepository,
	ucases.NewNetworkMetadataUCase,
)

var paymentStatisticsUCaseSet = wire.NewSet(
	repositories.NewPaymentStatisticsRepository,
	ucases.NewPaymentStatisticsCase,
)

// Init ucase
func InitializeBlockStateUCase(db *gorm.DB) interfaces.BlockStateUCase {
	wire.Build(blockStateUCaseSet)
	return nil
}

func InitializePaymentOrderUCase(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository, config *conf.Configuration) interfaces.PaymentOrderUCase {
	wire.Build(paymentOrderUCaseSet)
	return nil
}

func InitializeTokenTransferUCase(db *gorm.DB, config *conf.Configuration) interfaces.TokenTransferUCase {
	wire.Build(tokenTransferUCaseSet)
	return nil
}

func InitializePaymentEventHistoryUCase(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository, config *conf.Configuration) interfaces.PaymentEventHistoryUCase {
	wire.Build(paymentEventHistoryUCaseSet)
	return nil
}

func InitializePaymentWalletUCase(db *gorm.DB, config *conf.Configuration) interfaces.PaymentWalletUCase {
	wire.Build(paymentWalletUCaseSet)
	return nil
}

func InitializeUserWalletUCase(db *gorm.DB, config *conf.Configuration) interfaces.UserWalletUCase {
	wire.Build(userWalletUCaseSet)
	return nil
}

func InitializeNetworkMetadataUCase(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) interfaces.NetworkMetadataUCase {
	wire.Build(networkMetadataUCaseSet)
	return nil
}

func InitializePaymentStatisticsUCase(db *gorm.DB) interfaces.PaymentStatisticsUCase {
	wire.Build(paymentStatisticsUCaseSet)
	return nil
}
