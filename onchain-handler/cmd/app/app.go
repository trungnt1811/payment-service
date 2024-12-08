package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
	"github.com/genefriendway/onchain-handler/pkg/blockchain/eth"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/logger/zap"
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
	ethClientAvax := eth.NewClient(config.Blockchain.AvaxNetwork.AvaxRpcUrl)
	defer ethClientAvax.Close()

	// Initialize BSC client
	ethClientBsc := eth.NewClient(config.Blockchain.BscNetwork.BscRpcUrl)
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
	tokenTransferUCase := wire.InitializeTokenTransferUCase(db, ethClientAvax, config)
	networkMetadataUCase := wire.InitializeNetworkMetadataUCase(db, cacheRepository)

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
		config.Blockchain.AvaxNetwork.AvaxRpcUrl,
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		blockstateUcase,
		tokenTransferUCase,
		paymentOrderUCase,
		paymentWalletUCase,
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
		config.Blockchain.BscNetwork.BscRpcUrl,
		config.Blockchain.BscNetwork.BscUSDTContractAddress,
		blockstateUcase,
		tokenTransferUCase,
		paymentOrderUCase,
		paymentWalletUCase,
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
		paymentEventHistoryUCase,
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
		paymentEventHistoryUCase,
		paymentOrderQueue,
	)

	// Register routes
	routev1.RegisterRoutes(ctx, r, config, db, tokenTransferUCase, paymentOrderUCase, userWalletUCase, paymentWalletUCase, networkMetadataUCase)

	// Start server
	startServer(r, config)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// Helper Functions
func initializeLoggerAndMode(config *conf.Configuration) {
	// config logger config. Can be replaced with any logger config
	// Create a logger with default configuration
	factory := &zap.ZapLoggerFactory{}
	logger, err := factory.CreateLogger(nil)
	if err != nil {
		panic(err)
	}

	// Set service name and environment
	logger.SetServiceName(config.AppName)
	if config.Env == "PROD" {
		logger.SetConfigModeByCode(pkginterfaces.PRODUCTION_ENVIRONMENT_CODE_MODE)
		logger.SetLogLevel(pkginterfaces.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.SetConfigModeByCode(pkginterfaces.DEVELOPMENT_ENVIRONMENT_CODE_MODE)
		logger.SetLogLevel(pkginterfaces.DebugLevel)
	}

	// Use the logger
	pkglogger.InitLogger(logger)
	logger.Info("Application started")
	logger.WithFields(map[string]interface{}{
		"appName": config.AppName,
		"env":     config.Env,
	}).Info("Starting application")
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
		string(network),
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
	rpcURL, usdtContractAddress string,
	blockstateUcase interfaces.BlockStateUCase,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
) {
	latestBlockWorker := workers.NewLatestBlockWorker(cacheRepository, blockstateUcase, ethClient, network)
	go latestBlockWorker.Start(ctx)

	expiredOrderCatchupWorker := workers.NewExpiredOrderCatchupWorker(
		config,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		cacheRepository,
		usdtContractAddress,
		ethClient,
		network,
	)
	go expiredOrderCatchupWorker.Start(ctx)

	// Start payment wallet balance worker
	tokenSymbol, err := config.GetTokenSymbol(usdtContractAddress)
	if err != nil {
		pkglogger.GetLogger().Fatalf("Failed to get token symbol: %v", err)
	}
	paymentWalletBalanceWorker := workers.NewPaymentWalletBalanceWorker(
		paymentWalletUCase, cacheRepository, network, rpcURL, usdtContractAddress, tokenSymbol,
	)
	go paymentWalletBalanceWorker.Start(ctx)

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
		config.PaymentGateway.PrivateKeyMasterWallet,
		config.Wallet.Mnemonic,
		config.Wallet.Passphrase,
		config.Wallet.Salt,
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
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
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

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

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
