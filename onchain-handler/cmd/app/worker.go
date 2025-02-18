package app

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/infra/set"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/listeners"
	"github.com/genefriendway/onchain-handler/internal/workers"
	pkgblockchain "github.com/genefriendway/onchain-handler/pkg/blockchain"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/providers"
)

func RunWorkers(
	ctx context.Context,
	db *gorm.DB,
	config *conf.Configuration,
	cacheRepository infrainterfaces.CacheRepository,
	blockstateUcase interfaces.BlockStateUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
) {
	// Initialize payment order set
	paymentOrderSet := initializePaymentOrderSet(ctx, paymentOrderUCase.GetActivePaymentOrders)

	// Initialize AVAX C-Chain client
	avaxRpcUrls, err := conf.GetRpcUrls(constants.AvaxCChain)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to get AVAX C-Chain RPC URLs: %v", err)
	}
	ethClientAvax, err := providers.ProvideEthClient(constants.AvaxCChain, avaxRpcUrls)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to initialize AVAX C-Chain client: %v", err)
	}
	defer ethClientAvax.Close()

	// Initialize BSC client
	bscRpcUrls, err := conf.GetRpcUrls(constants.Bsc)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to get BSC RPC URLs: %v", err)
	}
	ethClientBsc, err := providers.ProvideEthClient(constants.Bsc, bscRpcUrls)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to initialize BSC client: %v", err)
	}
	defer ethClientBsc.Close()

	// Persist token decimals to cache
	persistTokenDecimalsToCache(
		ctx, ethClientAvax, config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress, constants.AvaxCChain, cacheRepository,
	)
	persistTokenDecimalsToCache(
		ctx, ethClientBsc, config.Blockchain.BscNetwork.BscUSDTContractAddress, constants.Bsc, cacheRepository,
	)

	// Start order clean worker
	releaseWalletWorker := workers.NewOrderCleanWorker(paymentOrderUCase)
	go releaseWalletWorker.Start(ctx)

	// Start AVAX workers
	startWorkers(
		ctx,
		config,
		cacheRepository,
		ethClientAvax,
		constants.AvaxCChain,
		uint64(config.Blockchain.AvaxNetwork.AvaxChainID),
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		blockstateUcase,
		tokenTransferUCase,
		paymentOrderUCase,
		paymentWalletUCase,
		paymentStatisticsUCase,
		paymentEventHistoryUCase,
	)

	// Start BSC workers
	startWorkers(
		ctx,
		config,
		cacheRepository,
		ethClientBsc,
		constants.Bsc,
		uint64(config.Blockchain.BscNetwork.BscChainID),
		config.Blockchain.BscNetwork.BscUSDTContractAddress,
		blockstateUcase,
		tokenTransferUCase,
		paymentOrderUCase,
		paymentWalletUCase,
		paymentStatisticsUCase,
		paymentEventHistoryUCase,
	)

	// Start AVAX event listeners
	startEventListeners(
		ctx,
		ethClientAvax,
		constants.AvaxCChain,
		config.Blockchain.AvaxNetwork.AvaxStartBlockListener,
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		cacheRepository,
		blockstateUcase,
		paymentOrderUCase,
		paymentStatisticsUCase,
		paymentEventHistoryUCase,
		paymentWalletUCase,
		paymentOrderSet,
	)

	// Start BSC event listeners
	startEventListeners(
		ctx,
		ethClientBsc,
		constants.Bsc,
		config.Blockchain.BscNetwork.BscStartBlockListener,
		config.Blockchain.BscNetwork.BscUSDTContractAddress,
		cacheRepository,
		blockstateUcase,
		paymentOrderUCase,
		paymentStatisticsUCase,
		paymentEventHistoryUCase,
		paymentWalletUCase,
		paymentOrderSet,
	)
}

// initializePaymentOrderSet initializes the payment order set with a unique key function.
func initializePaymentOrderSet(
	ctx context.Context,
	loader func(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error),
) infrainterfaces.Set[dto.PaymentOrderDTO] {
	// Key function to uniquely identify each PaymentOrderDTO
	keyFunc := func(order dto.PaymentOrderDTO) string {
		return order.PaymentAddress + "_" + order.Symbol
	}

	// Create the order set
	paymentOrderSet, err := set.NewSet(ctx, keyFunc, loader)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Create payment order set error: %v", err)
	}
	return paymentOrderSet
}

// persistTokenDecimalsToCache fetches token decimals from the blockchain and persists them to the cache
func persistTokenDecimalsToCache(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	tokenContractAddress string,
	network constants.NetworkType,
	cacheRepo infrainterfaces.CacheRepository,
) {
	_, err := pkgblockchain.FetchTokenDecimals(
		ctx,
		ethClient,
		tokenContractAddress,
		network.String(),
		cacheRepo,
	)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to persist token decimals to cache: %v", err)
	}
}

// startWorkers starts the workers for the given network
func startWorkers(
	ctx context.Context,
	config *conf.Configuration,
	cacheRepository infrainterfaces.CacheRepository,
	ethClient pkginterfaces.Client,
	network constants.NetworkType,
	chainID uint64,
	usdtContractAddress string,
	blockstateUcase interfaces.BlockStateUCase,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
) {
	latestBlockWorker := workers.NewLatestBlockWorker(cacheRepository, blockstateUcase, ethClient, network)
	go latestBlockWorker.Start(ctx)

	expiredOrderCatchupWorker := workers.NewExpiredOrderCatchupWorker(
		paymentOrderUCase,
		paymentEventHistoryUCase,
		paymentStatisticsUCase,
		paymentWalletUCase,
		cacheRepository,
		usdtContractAddress,
		ethClient,
		network,
	)
	go expiredOrderCatchupWorker.Start(ctx)

	// Start payment wallet withdraw worker
	paymentWalletWithdrawWorker := workers.NewPaymentWalletWithdrawWorker(
		ctx,
		ethClient,
		network,
		chainID,
		cacheRepository,
		tokenTransferUCase,
		paymentWalletUCase,
		usdtContractAddress,
		config.PaymentGateway.MasterWalletAddress,
		config.Wallet.Mnemonic,
		config.Wallet.Passphrase,
		config.Wallet.Salt,
		conf.GetGasBufferMultiplier(),
		config.PaymentGateway.WithdrawWorkerInterval,
	)
	go paymentWalletWithdrawWorker.Start(ctx)
}

// startEventListeners starts the event listeners for the given network
func startEventListeners(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	network constants.NetworkType,
	startBlockListener uint64,
	usdtContractAddress string,
	cacheRepository infrainterfaces.CacheRepository,
	blockstateUcase interfaces.BlockStateUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	paymentOrderSet infrainterfaces.Set[dto.PaymentOrderDTO],
) {
	baseEventListener := listeners.NewBaseEventListener(
		ethClient,
		network,
		cacheRepository,
		blockstateUcase,
		&startBlockListener,
	)

	tokenTransferListener, err := listeners.NewTokenTransferListener(
		ctx,
		cacheRepository,
		baseEventListener,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		paymentStatisticsUCase,
		paymentWalletUCase,
		network,
		usdtContractAddress,
		paymentOrderSet,
	)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to initialize tokenTransferListener: %v", err)
	}

	tokenTransferListener.Register(ctx)
	go func() {
		if err := baseEventListener.RunListener(ctx); err != nil {
			pkglogger.GetLogger().Errorf("Error running event listeners: %v", err)
		}
	}()
}
