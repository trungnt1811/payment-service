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

// Repo set
var paymentWalletRepoSet = wire.NewSet(
	payment_wallet.NewPaymentWalletRepository,
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

// Init repo
func InitializePaymentWalletRepository(db *gorm.DB, config *conf.Configuration) (interfaces.PaymentWalletRepository, error) {
	wire.Build(paymentWalletRepoSet)
	return nil, nil
}
