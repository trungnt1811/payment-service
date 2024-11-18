//go:build wireinject

package wire

import (
	"github.com/genefriendway/onchain-handler/conf"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/adapters/repositories"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/ucases"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
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
	repositories.NewPaymentWalletRepository,
	repositories.NewBlockstateRepository,
	ucases.NewPaymentOrderUCase,
)

var tokenTransferUCaseSet = wire.NewSet(
	repositories.NewTokenTransferRepository,
	ucases.NewTokenTransferUCase,
)

var paymentEventHistoryUCaseSet = wire.NewSet(
	repositories.NewPaymentEventHistoryRepository,
	ucases.NewPaymentEventHistoryUCase,
)

var paymentWalletUCaseSet = wire.NewSet(
	repositories.NewPaymentWalletRepository,
	ucases.NewPaymentWalletUCase,
)

var userWalletUCaseSet = wire.NewSet(
	repositories.NewUserWalletRepository,
	ucases.NewUserWalletUCase,
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

func InitializeTokenTransferUCase(db *gorm.DB, ethClient pkginterfaces.Client, config *conf.Configuration) interfaces.TokenTransferUCase {
	wire.Build(tokenTransferUCaseSet)
	return nil
}

func InitializePaymentEventHistoryUCase(db *gorm.DB) interfaces.PaymentEventHistoryUCase {
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
