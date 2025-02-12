package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/conf/database"
	"github.com/genefriendway/onchain-handler/constants"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/listeners"
	"github.com/genefriendway/onchain-handler/internal/middleware"
	routev1 "github.com/genefriendway/onchain-handler/internal/route"
	"github.com/genefriendway/onchain-handler/internal/workers"
	"github.com/genefriendway/onchain-handler/migrations"
	pkgblockchain "github.com/genefriendway/onchain-handler/pkg/blockchain"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/providers"
	"github.com/genefriendway/onchain-handler/wire"
)

func RunApp(config *conf.Configuration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger and environment settings
	initializeLoggerAndMode(config)

	// Initialize Gin router with middleware
	r := initializeRouter()

	// Initialize database connection
	db := database.DBConnWithLoglevel(logger.Info)
	if err := migrations.RunMigrations(db); err != nil {
		pkglogger.GetLogger().Fatalf("Failed to migrate database: %v", err)
	}

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

	// Initialize caching
	cacheRepository := providers.ProvideCacheRepository(ctx)

	// Persist token decimals to cache
	persistTokenDecimalsToCache(ctx, ethClientAvax, config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress, constants.AvaxCChain, cacheRepository)
	persistTokenDecimalsToCache(ctx, ethClientBsc, config.Blockchain.BscNetwork.BscUSDTContractAddress, constants.Bsc, cacheRepository)

	// Initialize use cases and queue
	blockstateUcase := wire.InitializeBlockStateUCase(db)
	paymentEventHistoryUCase := wire.InitializePaymentEventHistoryUCase(db, cacheRepository, config)
	paymentWalletUCase := wire.InitializePaymentWalletUCase(db, config)
	userWalletUCase := wire.InitializeUserWalletUCase(db, config)
	paymentOrderUCase := wire.InitializePaymentOrderUCase(db, cacheRepository, config)
	tokenTransferUCase := wire.InitializeTokenTransferUCase(db, config)
	networkMetadataUCase := wire.InitializeNetworkMetadataUCase(db, cacheRepository)
	paymentStatisticsUCase := wire.InitializePaymentStatisticsUCase(db)

	// Initialize payment order queue
	paymentOrderQueue := initializePaymentOrderQueue(ctx, paymentOrderUCase.GetActivePaymentOrders)

	// Initialize payment wallets
	initializePaymentWallets(ctx, config, paymentWalletUCase)

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
		config,
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
		paymentOrderQueue,
	)

	// Start BSC event listeners
	startEventListeners(
		ctx,
		config,
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
		paymentOrderQueue,
	)

	// Register routes
	routev1.RegisterRoutes(
		ctx,
		r,
		config,
		db,
		tokenTransferUCase,
		paymentOrderUCase,
		userWalletUCase,
		paymentWalletUCase,
		networkMetadataUCase,
		paymentStatisticsUCase,
	)

	// Start server
	startServer(r, config)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// Helper Functions
func initializeLoggerAndMode(config *conf.Configuration) {
	// Validate configuration
	if config == nil {
		panic("configuration cannot be nil")
	}

	// Determine the log level from the configuration
	var logLevel pkginterfaces.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logLevel = pkginterfaces.DebugLevel
		gin.SetMode(gin.DebugMode) // Development mode
	case "info":
		logLevel = pkginterfaces.InfoLevel
		gin.SetMode(gin.ReleaseMode) // Production mode
	default:
		// Default to info level if unspecified or invalid
		logLevel = pkginterfaces.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// Retrieve the initialized logger
	appLogger := pkglogger.GetLogger()

	// Set the log level in the logger package
	pkglogger.SetLogLevel(logLevel)

	// Log application startup details
	appLogger.Infof("Application '%s' started with log level '%s' in '%s' mode", config.AppName, logLevel, config.Env)

	// Log additional details for debugging
	if logLevel == pkginterfaces.DebugLevel {
		appLogger.Debug("Debugging mode enabled. Verbose logging is active.")
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(gin.Recovery())
	return r
}

func initializePaymentWallets(
	ctx context.Context,
	config *conf.Configuration,
	paymentWalletUCase interfaces.PaymentWalletUCase,
) {
	err := payment.InitPaymentWallets(
		ctx,
		config.Wallet.Mnemonic,
		config.Wallet.Passphrase,
		config.Wallet.Salt,
		config.PaymentGateway.InitWalletCount,
		paymentWalletUCase,
	)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Init payment wallets error: %v", err)
	}
}

func initializePaymentOrderQueue(
	ctx context.Context,
	loader func(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error),
) *queue.Queue[dto.PaymentOrderDTO] {
	paymentOrderQueue, err := queue.NewQueue(ctx, constants.MinQueueLimit, loader)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Create payment order queue error: %v", err)
	}
	return paymentOrderQueue
}

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
		config,
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
		config.GetGasBufferMultiplier(),
		config.PaymentGateway.WithdrawWorkerInterval,
	)
	go paymentWalletWithdrawWorker.Start(ctx)
}

func startEventListeners(
	ctx context.Context,
	config *conf.Configuration,
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
	paymentOrderQueue *queue.Queue[dto.PaymentOrderDTO],
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
		config,
		cacheRepository,
		baseEventListener,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		paymentStatisticsUCase,
		paymentWalletUCase,
		network,
		usdtContractAddress,
		paymentOrderQueue,
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

func startServer(
	r *gin.Engine,
	config *conf.Configuration,
) {
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("%s is still alive", config.AppName),
		})
	})

	if config.Env != "PROD" {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	go func() {
		if err := r.Run(fmt.Sprintf("0.0.0.0:%v", config.AppPort)); err != nil {
			pkglogger.GetLogger().Fatalf("Failed to run gin router: %v", err)
		}
	}()
}

func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	pkglogger.GetLogger().Debug("Shutting down gracefully...")
	cancel()
}
