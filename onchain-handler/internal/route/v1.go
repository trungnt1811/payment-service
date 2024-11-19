package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/handlers"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	transferUCase interfaces.TokenTransferUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	userWalletUCase interfaces.UserWalletUCase,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferHandler := handlers.NewTokenTransferHandler(transferUCase, config)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
	appRouter.GET("/token-transfer/histories", transferHandler.GetTokenTransferHistories)

	// SECTION: payment order
	paymentOrderHandler := handlers.NewPaymentOrderHandler(paymentOrderUCase, config)
	appRouter.POST("/payment-orders", paymentOrderHandler.CreateOrders)
	appRouter.GET("/payment-orders", paymentOrderHandler.GetPaymentOrderHistories)
	appRouter.GET("/payment-orders/:id", paymentOrderHandler.GetPaymentOrderByID)

	// SECTION: user wallet
	userWalletHander := handlers.NewUserWalletHandler(userWalletUCase, config)
	appRouter.POST("/user-wallets", userWalletHander.CreateUserWallets)
	appRouter.GET("/user-wallets", userWalletHander.GetUserWallets)
}
