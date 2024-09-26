package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/blockchain"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/module/blockstate"
	"github.com/genefriendway/onchain-handler/internal/module/membership"
	"github.com/genefriendway/onchain-handler/internal/module/transfer"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

func RegisterRoutes(r *gin.Engine, config *conf.Configuration, db *gorm.DB, ethClient *ethclient.Client, ctx context.Context) {
	v1 := r.Group("/api/v1")
	appRouter := v1.Group("")

	// SECTION: reward tokens
	transferRepository := transfer.NewTransferRepository(db)
	transferUCase := transfer.NewtTransferUCase(transferRepository, ethClient, config)
	transferHandler := transfer.NewTransferHandler(transferUCase)
	appRouter.POST("/transfer", transferHandler.Transfer)

	// SECTION: membership purchase
	membershipRepository := membership.NewMembershipRepository(db)
	membershipUCase := membership.NewMembershipUCase(membershipRepository)
	membershipHandler := membership.NewMembershipHandler(membershipUCase)
	appRouter.GET("/membership/events", membershipHandler.GetMembershipEventsByOrderIDs)

	// SECTION: events listener
	membershipEventListener, err := blockchain.NewMembershipEventListener(
		ethClient,
		config.Blockchain.MembershipContractAddress,
		membershipRepository,
		blockstate.NewBlockstateRepository(db),
		&config.Blockchain.StartBlockListener,
	)
	if err != nil {
		log.LG.Errorf("Failed to initialize MembershipEventListener: %v", err)
		return
	}
	go func() {
		if err := membershipEventListener.RunListener(ctx); err != nil {
			log.LG.Errorf("Error running MembershipEventListener: %v", err)
		}
	}()
}
