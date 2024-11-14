package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/conf/database"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/listeners"
	"github.com/genefriendway/onchain-handler/internal/middleware"
	routev1 "github.com/genefriendway/onchain-handler/internal/route"
	"github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/internal/workers"
	"github.com/genefriendway/onchain-handler/log"
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

	// Initialize AVAX C-Chain client
	ethClientAvax := initializeEthClient(config.Blockchain.AvaxNetwork.AvaxRpcUrl)
	defer ethClientAvax.Close()

	// Initialize BSC client
	ethClientBsc := initializeEthClient(config.Blockchain.BscNetwork.BscRpcUrl)
	defer ethClientBsc.Close()

	// Initialize caching
	cacheRepository := initializeCache(ctx)

	// Initialize use cases and queue
	blockstateUcase := wire.InitializeBlockStateUCase(db)
	paymentEventHistoryUCase := wire.InitializePaymentEventHistoryUCase(db)
	paymentWalletUCase := wire.InitializePaymentWalletUCase(db, config)
	userWalletUCase := wire.InitializeUserWalletUCase(db, config)
	paymentOrderUCase := wire.InitializePaymentOrderUCase(db, cacheRepository, config)
	transferUCase := wire.InitializeTokenTransferUCase(db, ethClientAvax, config)

	// Initialize AVAX C-Chain payment order queue
	paymentOrderQueueOnAvax := initializePaymentOrderQueue(ctx, paymentOrderUCase.GetActivePaymentOrdersOnAvax)

	// Initialize BSC payment order queue
	paymentOrderQueueOnBsc := initializePaymentOrderQueue(ctx, paymentOrderUCase.GetActivePaymentOrdersOnBsc)

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
		config.Blockchain.AvaxNetwork.AvaxUSDTContractAddress,
		blockstateUcase,
		paymentOrderUCase,
		paymentEventHistoryUCase,
	)

	// Start BSC workers
	startWorkers(
		ctx,
		config,
		cacheRepository,
		ethClientBsc,
		constants.Bsc,
		config.Blockchain.BscNetwork.BscUSDTContractAddress,
		blockstateUcase,
		paymentOrderUCase,
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
		paymentOrderQueueOnAvax,
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
		paymentOrderQueueOnBsc,
	)

	// Register routes
	routev1.RegisterRoutes(ctx, r, config, db, transferUCase, paymentOrderUCase, userWalletUCase, ethClientAvax)

	// Start server
	startServer(r, config)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// Helper Functions

func initializeLoggerAndMode(config *conf.Configuration) {
	if config.Env == "PROD" {
		log.InitZerologLogger(os.Stdout, zerolog.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.InitZerologLogger(os.Stdout, zerolog.DebugLevel)
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLogger(log.LG.Instance))
	r.Use(gin.Recovery())
	return r
}

func initializeEthClient(rpcUrl string) *ethclient.Client {
	ethClient, err := utils.InitEthClient(rpcUrl)
	if err != nil {
		log.LG.Fatalf("Failed to connect to ETH client: %v", err)
	}
	return ethClient
}

func initializeCache(ctx context.Context) infrainterfaces.CacheRepository {
	cacheClient := caching.NewGoCacheClient()
	return caching.NewCachingRepository(ctx, cacheClient)
}

func initializePaymentWallets(
	ctx context.Context,
	config *conf.Configuration,
	paymentWalletUCase interfaces.PaymentWalletUCase,
) {
	err := utils.InitPaymentWallets(ctx, config, paymentWalletUCase)
	if err != nil {
		log.LG.Fatalf("Init payment wallets error: %v", err)
	}
}

func initializePaymentOrderQueue(
	ctx context.Context,
	loader func(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error),
) *queue.Queue[dto.PaymentOrderDTO] {
	paymentOrderQueue, err := queue.NewQueue(ctx, constants.MinQueueLimit, loader)
	if err != nil {
		log.LG.Fatalf("Create payment order queue error: %v", err)
	}
	return paymentOrderQueue
}

func startWorkers(
	ctx context.Context,
	config *conf.Configuration,
	cacheRepository infrainterfaces.CacheRepository,
	ethClient *ethclient.Client,
	network constants.NetworkType,
	usdtContractAddress string,
	blockstateUcase interfaces.BlockStateUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
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
}

func startEventListeners(
	ctx context.Context,
	config *conf.Configuration,
	ethClient *ethclient.Client,
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
		baseEventListener,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		network,
		usdtContractAddress,
		paymentOrderQueue,
	)
	if err != nil {
		log.LG.Fatalf("Failed to initialize tokenTransferListener: %v", err)
	}

	tokenTransferListener.Register(ctx)
	go func() {
		if err := baseEventListener.RunListener(ctx); err != nil {
			log.LG.Errorf("Error running event listeners: %v", err)
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
			log.LG.Fatalf("Failed to run gin router: %v", err)
		}
	}()
}

func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	log.LG.Info("Shutting down gracefully...")
	cancel()
}
