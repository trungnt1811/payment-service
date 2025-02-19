package app

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/onchain-handler/conf"
	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/http/middleware"
	routev1 "github.com/genefriendway/onchain-handler/internal/delivery/http/route"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
)

func RunServer(
	ctx context.Context,
	db *gorm.DB,
	config *conf.Configuration,
	cacheRepository cachetypes.CacheRepository,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	tokenTransferUCase ucasetypes.TokenTransferUCase,
	networkMetadataUCase ucasetypes.NetworkMetadataUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
) {
	// Initialize Gin router with middleware
	r := initializeRouter()

	// Initialize payment wallets
	initializePaymentWallets(ctx, config, paymentWalletUCase)

	// Register routes
	routev1.RegisterRoutes(
		ctx,
		r,
		config,
		db,
		tokenTransferUCase,
		paymentOrderUCase,
		paymentWalletUCase,
		networkMetadataUCase,
		paymentStatisticsUCase,
	)

	// Start server
	startServer(r, config)
}

// Helper Functions
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
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
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
