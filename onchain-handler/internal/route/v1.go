package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	transfer "github.com/genefriendway/onchain-handler/internal/module/token_transfer"
)

func RegisterRoutes(r *gin.Engine, config *conf.Configuration, db *gorm.DB, ethClient *ethclient.Client, ctx context.Context) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferRepository := transfer.NewTransferRepository(db)
	transferUCase := transfer.NewTransferUCase(transferRepository, ethClient, config)
	transferHandler := transfer.NewTransferHandler(transferUCase)
	appRouter.POST("/transfer", transferHandler.Transfer)
}
