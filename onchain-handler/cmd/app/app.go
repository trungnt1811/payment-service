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

	// Initialize ETH client
	ethClient := initializeEthClient(config)
	defer ethClient.Close()

	// Initialize caching
	cacheRepository := initializeCache(ctx)

	// Initialize use cases and queue
	blockstateUcase, _ := wire.InitializeBlockStateUCase(db)
	paymentEventHistoryUCase, _ := wire.InitializePaymentEventHistoryUCase(db)
	paymentWalletUCase, _ := wire.InitializePaymentWalletUCase(db, config)
	userWalletUCase, _ := wire.InitializeUserWalletUCase(db, config)
	paymentOrderUCase, _ := wire.InitializePaymentOrderUCase(db, cacheRepository, config)
	transferUCase, _ := wire.InitializeTokenTransferUCase(db, ethClient, config)
	paymentOrderQueue := initializePaymentOrderQueue(ctx, paymentOrderUCase)

	// Initialize payment wallets
	initializePaymentWallets(ctx, config, paymentWalletUCase)

	// Start workers
	startWorkers(ctx, config, cacheRepository, ethClient, blockstateUcase, paymentOrderUCase, paymentEventHistoryUCase)

	// Start event listeners
	startEventListeners(ctx, config, ethClient, cacheRepository, blockstateUcase, paymentOrderUCase, paymentEventHistoryUCase, paymentOrderQueue)

	// Register routes
	routev1.RegisterRoutes(ctx, r, config, db, transferUCase, paymentOrderUCase, userWalletUCase, ethClient)

	// Start server
	startServer(r, config)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// Helper Functions

func initializeLoggerAndMode(config *conf.Configuration) {
	if config.Env == "PROD" {
		log.LG = log.NewZerologLogger(os.Stdout, zerolog.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.LG = log.NewZerologLogger(os.Stdout, zerolog.DebugLevel)
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLogger(log.LG.Instance))
	r.Use(gin.Recovery())
	return r
}

func initializeEthClient(config *conf.Configuration) *ethclient.Client {
	ethClient, err := utils.InitEthClient(config.Blockchain.RpcUrl)
	if err != nil {
		log.LG.Fatalf("Failed to connect to ETH client: %v", err)
	}
	return ethClient
}

func initializeCache(ctx context.Context) caching.CacheRepository {
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

func initializePaymentOrderQueue(ctx context.Context, paymentOrderUCase interfaces.PaymentOrderUCase) *queue.Queue[dto.PaymentOrderDTO] {
	paymentOrderQueue, err := queue.NewQueue[dto.PaymentOrderDTO](ctx, constants.MinQueueLimit, paymentOrderUCase.GetActivePaymentOrders)
	if err != nil {
		log.LG.Fatalf("Create payment order queue error: %v", err)
	}
	return paymentOrderQueue
}

func startWorkers(
	ctx context.Context,
	config *conf.Configuration,
	cacheRepository caching.CacheRepository,
	ethClient *ethclient.Client,
	blockstateUcase interfaces.BlockStateUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
) {
	latestBlockWorker := workers.NewLatestBlockWorker(cacheRepository, blockstateUcase, ethClient)
	go latestBlockWorker.Start(ctx)

	expiredOrderCatchupWorker := workers.NewExpiredOrderCatchupWorker(
		config,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		cacheRepository,
		config.Blockchain.SmartContract.USDTContractAddress,
		ethClient,
	)
	go expiredOrderCatchupWorker.Start(ctx)

	releaseWalletWorker := workers.NewOrderCleanWorker(paymentOrderUCase)
	go releaseWalletWorker.Start(ctx)
}

func startEventListeners(
	ctx context.Context,
	config *conf.Configuration,
	ethClient *ethclient.Client,
	cacheRepository caching.CacheRepository,
	blockstateUcase interfaces.BlockStateUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	paymentEventHistoryUCase interfaces.PaymentEventHistoryUCase,
	paymentOrderQueue *queue.Queue[dto.PaymentOrderDTO],
) {
	baseEventListener := listeners.NewBaseEventListener(
		ethClient,
		cacheRepository,
		blockstateUcase,
		&config.Blockchain.StartBlockListener,
	)

	tokenTransferListener, err := listeners.NewTokenTransferListener(
		ctx,
		config,
		baseEventListener,
		paymentOrderUCase,
		paymentEventHistoryUCase,
		config.Blockchain.SmartContract.USDTContractAddress,
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
