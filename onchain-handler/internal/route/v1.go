package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	paymentorder "github.com/genefriendway/onchain-handler/internal/module/payment_order"
	tokentransfer "github.com/genefriendway/onchain-handler/internal/module/token_transfer"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	transferUCase interfaces.TokenTransferUCase,
	paymentOrderUCase interfaces.PaymentOrderUCase,
	ethClient *ethclient.Client,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferHandler := tokentransfer.NewTokenTransferHandler(transferUCase, config)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
	appRouter.GET("/token-transfer/histories", transferHandler.GetTokenTransferHistories)

	// SECTION: payment order
	paymentOrderHandler := paymentorder.NewPaymentOrderHandler(paymentOrderUCase, config)
	appRouter.POST("/payment-orders", paymentOrderHandler.CreateOrders)
	appRouter.GET("payment-orders/histories", paymentOrderHandler.GetPaymentOrderHistories)
}
