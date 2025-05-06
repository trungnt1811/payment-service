package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"

	app "github.com/genefriendway/onchain-handler/cmd/app"
	"github.com/genefriendway/onchain-handler/conf"
	_ "github.com/genefriendway/onchain-handler/docs"
	"github.com/genefriendway/onchain-handler/internal/adapters/database/postgres"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	loggertypes "github.com/genefriendway/onchain-handler/pkg/logger/types"
	"github.com/genefriendway/onchain-handler/wire"
	"github.com/genefriendway/onchain-handler/wire/providers"
)

func main() {
	// Initialize the application context and defer the cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load the application configuration
	config := conf.GetConfiguration()

	// Initialize database connection
	db := providers.ProvideDBConnection()

	// TODO: remove this migate function later
	basePath := "internal/adapters/database/postgres/scripts"
	log.Printf("Running database migrations from: %s", basePath)
	if err := postgres.RunMigrations(db, basePath); err != nil {
		log.Panicf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully.")

	// Initialize the logger and set the application mode
	initializeLoggerAndMode(config)

	// Initialize the cache repository
	cacheRepository := providers.ProvideCacheRepository(ctx)

	// Initialize payment order set
	paymentOrderSet := providers.ProvidePaymentOrderSet(ctx)

	// Initialize use cases
	ucases := wire.InitializeUseCases(db, cacheRepository, paymentOrderSet)

	if config.WorkerEnabled {
		// Run the application workers
		app.RunWorkers(
			ctx, db, config, cacheRepository,
			ucases.BlockStateUCase,
			ucases.PaymentEventHistoryUCase,
			ucases.PaymentOrderUCase,
			ucases.TokenTransferUCase,
			ucases.PaymentWalletUCase,
			ucases.PaymentStatisticsUCase,
			paymentOrderSet,
		)
	}

	// Run the application server
	app.RunServer(
		ctx, db, config, cacheRepository,
		ucases.PaymentWalletUCase,
		ucases.PaymentOrderUCase,
		ucases.TokenTransferUCase,
		ucases.MetadataUCase,
		ucases.PaymentStatisticsUCase,
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
	var logLevel loggertypes.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logLevel = loggertypes.DebugLevel
		gin.SetMode(gin.DebugMode) // Development mode
	case "info":
		logLevel = loggertypes.InfoLevel
		gin.SetMode(gin.ReleaseMode) // Production mode
	default:
		// Default to info level if unspecified or invalid
		logLevel = loggertypes.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// Retrieve the initialized logger
	appLogger := pkglogger.GetLogger()

	// Set the log level in the logger package
	pkglogger.SetLogLevel(logLevel)

	// Log application startup details
	appLogger.Infof("Application '%s' started with log level '%s' in '%s' mode", config.AppName, logLevel, config.Env)

	// Log additional details for debugging
	if logLevel == loggertypes.DebugLevel {
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
