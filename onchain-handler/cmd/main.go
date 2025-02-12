package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"

	app "github.com/genefriendway/onchain-handler/cmd/app"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/conf/database"
	"github.com/genefriendway/onchain-handler/constants"
	_ "github.com/genefriendway/onchain-handler/docs"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/migrations"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/providers"
	"github.com/genefriendway/onchain-handler/wire"
)

func main() {
	// Initialize the application context and defer the cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load the application configuration
	config := conf.GetConfiguration()

	// Initialize database connection
	db := database.DBConnWithLoglevel(logger.Info)
	if err := migrations.RunMigrations(db); err != nil {
		pkglogger.GetLogger().Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize the logger and set the application mode
	initializeLoggerAndMode(config)

	// Initialize the cache repository
	cacheRepository := providers.ProvideCacheRepository(ctx)

	// Initialize use cases
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

	// Run the application worker
	app.RunWorker(
		ctx, db, config, paymentOrderQueue, cacheRepository,
		blockstateUcase, paymentEventHistoryUCase, paymentOrderUCase, tokenTransferUCase, paymentWalletUCase, paymentStatisticsUCase,
	)

	// Run the application server
	app.RunServer(
		ctx, db, config, cacheRepository,
		paymentWalletUCase, userWalletUCase, paymentOrderUCase, tokenTransferUCase, networkMetadataUCase, paymentStatisticsUCase,
	)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// initializeLoggerAndMode initializes the logger and sets the application mode based on the configuration
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

// waitForShutdownSignal waits for a shutdown signal and cancels the context
func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	pkglogger.GetLogger().Debug("Shutting down gracefully...")
	cancel()
}

// initializePaymentWallets initializes the payment wallets
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
