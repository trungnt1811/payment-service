package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/module/token_transfer"
)

func RegisterRoutes(r *gin.Engine, config *conf.Configuration, db *gorm.DB, ethClient *ethclient.Client, ctx context.Context) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: tokens transfer
	transferRepository := token_transfer.NewTokenTransferRepository(db)
	transferUCase := token_transfer.NewTokenTransferUCase(transferRepository, ethClient, config)
	transferHandler := token_transfer.NewTokenTransferHandler(transferUCase)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
}
