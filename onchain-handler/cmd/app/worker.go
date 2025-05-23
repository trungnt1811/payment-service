package app

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	settypes "github.com/genefriendway/onchain-handler/internal/adapters/orderset/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	"github.com/genefriendway/onchain-handler/internal/listeners"
	"github.com/genefriendway/onchain-handler/internal/wire/instances"
	"github.com/genefriendway/onchain-handler/internal/workers"
	pkgblockchain "github.com/genefriendway/onchain-handler/pkg/blockchain"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
)

func RunWorkers(
	ctx context.Context,
	db *gorm.DB,
	config *conf.Configuration,
	cacheRepository cachetypes.CacheRepository,
	blockStateUCase ucasetypes.BlockStateUCase,
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	tokenTransferUCase ucasetypes.TokenTransferUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
	paymentOrderSet settypes.Set[dto.PaymentOrderDTO],
) {
	// Initialize AVAX C-Chain client
	avaxRPCUrls, err := conf.GetRPCUrls(constants.AvaxCChain)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to get AVAX C-Chain RPC URLs: %v", err)
	}
	ethClientAvax, err := instances.ETHClientInstance(constants.AvaxCChain, avaxRPCUrls)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to initialize AVAX C-Chain client: %v", err)
	}
	defer ethClientAvax.Close()

	// Initialize BSC client
	bscRPCUrls, err := conf.GetRPCUrls(constants.Bsc)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to get BSC RPC URLs: %v", err)
	}
	ethClientBsc, err := instances.ETHClientInstance(constants.Bsc, bscRPCUrls)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to initialize BSC client: %v", err)
	}
	defer ethClientBsc.Close()

	// Persist token decimals to cache
	type tokenInfo struct {
		Client          clienttypes.Client
		ContractAddress string
		Network         constants.NetworkType
	}
	tokens := []tokenInfo{
		{ethClientAvax, config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress, constants.AvaxCChain},
		{ethClientAvax, config.Blockchain.AvaxNetwork.AvaxUSDCContractAddress, constants.AvaxCChain},
		{ethClientBsc, config.Blockchain.BscNetwork.BscUSDTContractAddress, constants.Bsc},
		{ethClientBsc, config.Blockchain.BscNetwork.BscUSDCContractAddress, constants.Bsc},
	}
	for _, token := range tokens {
		persistTokenDecimalsToCache(ctx, token.Client, token.ContractAddress, token.Network, cacheRepository)
	}

	// Start order clean worker
	releaseWalletWorker := workers.NewOrderCleanWorker(paymentOrderUCase, paymentOrderSet)
	go releaseWalletWorker.Start(ctx)

	// Token contract addresses for AVAX and BSC
	tokenBSCContractAddresses := []string{
		config.Blockchain.BscNetwork.BscUSDTContractAddress,
		config.Blockchain.BscNetwork.BscUSDCContractAddress,
	}

	tokenAVAXContractAddresses := []string{
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		config.Blockchain.AvaxNetwork.AvaxUSDCContractAddress,
	}

	// Start AVAX workers
	startWorkers(
		ctx,
		config,
		cacheRepository,
		ethClientAvax,
		constants.AvaxCChain,
		uint64(config.Blockchain.AvaxNetwork.AvaxChainID),
		tokenAVAXContractAddresses,
		blockStateUCase,
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
		tokenBSCContractAddresses,
		blockStateUCase,
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
		tokenAVAXContractAddresses,
		cacheRepository,
		blockStateUCase,
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
		tokenBSCContractAddresses,
		cacheRepository,
		blockStateUCase,
		paymentOrderUCase,
		paymentStatisticsUCase,
		paymentEventHistoryUCase,
		paymentWalletUCase,
		paymentOrderSet,
	)
}

// persistTokenDecimalsToCache fetches token decimals from the blockchain and persists them to the cache
func persistTokenDecimalsToCache(
	ctx context.Context,
	ethClient clienttypes.Client,
	tokenContractAddress string,
	network constants.NetworkType,
	cacheRepo cachetypes.CacheRepository,
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
	cacheRepository cachetypes.CacheRepository,
	ethClient clienttypes.Client,
	network constants.NetworkType,
	chainID uint64,
	tokenContractAddresses []string,
	blockStateUCase ucasetypes.BlockStateUCase,
	tokenTransferUCase ucasetypes.TokenTransferUCase,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase,
) {
	latestBlockWorker := workers.NewLatestBlockWorker(blockStateUCase, ethClient, network)
	go latestBlockWorker.Start(ctx)

	expiredOrderCatchupWorker := workers.NewExpiredOrderCatchupWorker(
		paymentOrderUCase,
		paymentEventHistoryUCase,
		paymentStatisticsUCase,
		paymentWalletUCase,
		blockStateUCase,
		cacheRepository,
		tokenContractAddresses,
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
		tokenContractAddresses,
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
	ethClient clienttypes.Client,
	network constants.NetworkType,
	startBlockListener uint64,
	tokenContractAddresses []string,
	cacheRepository cachetypes.CacheRepository,
	blockstateUcase ucasetypes.BlockStateUCase,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
	paymentEventHistoryUCase ucasetypes.PaymentEventHistoryUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	paymentOrderSet settypes.Set[dto.PaymentOrderDTO],
) {
	baseEventListener := listeners.NewBaseEventListener(
		ethClient,
		network,
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
		tokenContractAddresses,
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
