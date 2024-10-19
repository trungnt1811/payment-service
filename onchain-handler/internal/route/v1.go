package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/infra/caching"
	paymentorder "github.com/genefriendway/onchain-handler/internal/module/payment_order"
	tokentransfer "github.com/genefriendway/onchain-handler/internal/module/token_transfer"
	"github.com/genefriendway/onchain-handler/wire"
)

func RegisterRoutes(
	ctx context.Context,
	r *gin.Engine,
	config *conf.Configuration,
	db *gorm.DB,
	cacheRepository caching.CacheRepository,
	ethClient *ethclient.Client,
) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferUCase, _ := wire.InitializeTokenTransferUCase(db, ethClient, config)
	transferHandler := tokentransfer.NewTokenTransferHandler(transferUCase, config)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
	appRouter.GET("/token-transfer/histories", transferHandler.GetTokenTransferHistories)

	// SECTION: payment order
	paymentOrderUCase, _ := wire.InitializePaymentOrderUCase(db, cacheRepository, config)
	paymentOrderHandler := paymentorder.NewPaymentOrderHandler(paymentOrderUCase, config)
	appRouter.POST("/payment-orders", paymentOrderHandler.CreateOrders)
}
