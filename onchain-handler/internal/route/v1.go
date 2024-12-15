package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/handlers"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/middleware"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	userWalletUCase interfaces.UserWalletUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	networkMetadataUCase interfaces.NetworkMetadataUCase,
	paymentStatisticsUCase interfaces.PaymentStatisticsUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferHandler := handlers.NewTokenTransferHandler(tokenTransferUCase, config)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
	appRouter.GET("/token-transfers", transferHandler.GetTokenTransferHistories)

	// SECTION: payment order
	paymentOrderHandler := handlers.NewPaymentOrderHandler(paymentOrderUCase, config)
	appRouter.POST("/payment-orders", middleware.ValidateVendorID(), paymentOrderHandler.CreateOrders)
	appRouter.GET("/payment-orders", middleware.ValidateVendorID(), paymentOrderHandler.GetPaymentOrders)
	appRouter.GET("/payment-order/:request_id", paymentOrderHandler.GetPaymentOrderByRequestID)
	appRouter.PUT("/payment-order/network", paymentOrderHandler.UpdatePaymentOrderNetwork)

	// SECTION: user wallet
	userWalletHander := handlers.NewUserWalletHandler(userWalletUCase, config)
	appRouter.POST("/user-wallets", userWalletHander.CreateUserWallets)
	appRouter.GET("/user-wallets", userWalletHander.GetUserWallets)

	// SECTION: payment wallet
	paymentWalletHander := handlers.NewPaymentWalletHandler(paymentWalletUCase)
	appRouter.GET("/payment-wallet/:address", paymentWalletHander.GetPaymentWalletByAddress)
	appRouter.GET("/payment-wallets/balances", paymentWalletHander.GetPaymentWalletsWithBalances)

	// SECTION: metadata
	metadataHandler := handlers.NewMetadataHandler(networkMetadataUCase)
	appRouter.GET("/metadata/networks", metadataHandler.GetNetworksMetadata)

	// SECTION: payment statistics
	paymentStatisticsHandler := handlers.NewPaymentStatisticsHandler(paymentStatisticsUCase)
	appRouter.GET("payment-statistics", paymentStatisticsHandler.GetPaymentStatistics)
}
