//go:build wireinject

package wire

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/module/block_state"
	"github.com/genefriendway/onchain-handler/internal/module/payment_event_history"
	"github.com/genefriendway/onchain-handler/internal/module/payment_order"
	"github.com/genefriendway/onchain-handler/internal/module/payment_wallet"
	"github.com/genefriendway/onchain-handler/internal/module/token_transfer"
	"github.com/genefriendway/onchain-handler/internal/module/user_wallet"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var blockStateUCaseSet = wire.NewSet(
	block_state.NewBlockstateRepository,
	block_state.NewBlockStateUCase,
)

var paymentOrderUCaseSet = wire.NewSet(
	payment_order.NewPaymentOrderRepository,
	payment_wallet.NewPaymentWalletRepository,
	block_state.NewBlockstateRepository,
	payment_order.NewPaymentOrderUCase,
)

var tokenTransferUCaseSet = wire.NewSet(
	token_transfer.NewTokenTransferRepository,
	token_transfer.NewTokenTransferUCase,
)

var paymentEventHistoryUCaseSet = wire.NewSet(
	payment_event_history.NewPaymentEventHistoryRepository,
	payment_event_history.NewPaymentEventHistoryUCase,
)

var paymentWalletUCaseSet = wire.NewSet(
	payment_wallet.NewPaymentWalletRepository,
	payment_wallet.NewPaymentWalletUCase,
)

var userWalletUCaseSet = wire.NewSet(
	user_wallet.NewUserWalletRepository,
	user_wallet.NewUserWalletUCase,
)

// Init ucase
func InitializeBlockStateUCase(db *gorm.DB) (interfaces.BlockStateUCase, error) {
	wire.Build(blockStateUCaseSet)
	return nil, nil
}

func InitializePaymentOrderUCase(db *gorm.DB, cacheRepo caching.CacheRepository, config *conf.Configuration) (interfaces.PaymentOrderUCase, error) {
	wire.Build(paymentOrderUCaseSet)
	return nil, nil
}

func InitializeTokenTransferUCase(db *gorm.DB, ethClient *ethclient.Client, config *conf.Configuration) (interfaces.TokenTransferUCase, error) {
	wire.Build(tokenTransferUCaseSet)
	return nil, nil
}

func InitializePaymentEventHistoryUCase(db *gorm.DB) (interfaces.PaymentEventHistoryUCase, error) {
	wire.Build(paymentEventHistoryUCaseSet)
	return nil, nil
}

func InitializePaymentWalletUCase(db *gorm.DB, config *conf.Configuration) (interfaces.PaymentWalletUCase, error) {
	wire.Build(paymentWalletUCaseSet)
	return nil, nil
}

func InitializeUserWalletUCase(db *gorm.DB, config *conf.Configuration) (interfaces.UserWalletUCase, error) {
	wire.Build(userWalletUCaseSet)
	return nil, nil
}
