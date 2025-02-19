package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/delivery/http/handlers"
	"github.com/genefriendway/onchain-handler/internal/delivery/http/middleware"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	tokenTransferUCase ucasetypes.TokenTransferUCase,
	paymentOrderUCase ucasetypes.PaymentOrderUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	networkMetadataUCase ucasetypes.NetworkMetadataUCase,
	paymentStatisticsUCase ucasetypes.PaymentStatisticsUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferHandler := handlers.NewTokenTransferHandler(paymentWalletUCase, tokenTransferUCase, config)
	appRouter.GET("/token-transfers", transferHandler.GetTokenTransferHistories)
	appRouter.GET("/withdraws", transferHandler.GetWithdrawHistories)

	// SECTION: payment order
	paymentOrderHandler := handlers.NewPaymentOrderHandler(paymentOrderUCase)
	appRouter.POST("/payment-orders", middleware.ValidateVendorID(), paymentOrderHandler.CreateOrders)
	appRouter.GET("/payment-orders", middleware.ValidateVendorID(), paymentOrderHandler.GetPaymentOrders)
	appRouter.GET("/payment-order/:request_id", paymentOrderHandler.GetPaymentOrderByRequestID)
	appRouter.PUT("/payment-order/network", paymentOrderHandler.UpdatePaymentOrderNetwork)

	// SECTION: payment wallet
	paymentWalletHander := handlers.NewPaymentWalletHandler(paymentWalletUCase, config)
	appRouter.GET("/payment-wallet/:address", paymentWalletHander.GetPaymentWalletByAddress)
	appRouter.GET("/payment-wallets/balances", paymentWalletHander.GetPaymentWalletsWithBalances)
	appRouter.GET("/payment-wallets/receiving-address", paymentWalletHander.GetReceivingWalletAddress)
	appRouter.PUT("payment-wallets/balance/sync", paymentWalletHander.SyncPaymentWalletBalance)

	// SECTION: metadata
	metadataHandler := handlers.NewMetadataHandler(networkMetadataUCase)
	appRouter.GET("/metadata/networks", metadataHandler.GetNetworksMetadata)

	// SECTION: payment statistics
	paymentStatisticsHandler := handlers.NewPaymentStatisticsHandler(paymentStatisticsUCase)
	appRouter.GET("payment-statistics", paymentStatisticsHandler.GetPaymentStatistics)
}
