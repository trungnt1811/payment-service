package route

import (
	"context"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/blockchain/listener"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/infra/queue"
	"github.com/genefriendway/onchain-handler/internal/dto"
	blockstate "github.com/genefriendway/onchain-handler/internal/module/block_state"
	paymentorder "github.com/genefriendway/onchain-handler/internal/module/payment_order"
	paymentwallet "github.com/genefriendway/onchain-handler/internal/module/payment_wallet"
	tokentransfer "github.com/genefriendway/onchain-handler/internal/module/token_transfer"
	"github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
	"github.com/genefriendway/onchain-handler/worker"
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
	transferRepository := tokentransfer.NewTokenTransferRepository(db)
	transferUCase := tokentransfer.NewTokenTransferUCase(transferRepository, ethClient, config)
	transferHandler := tokentransfer.NewTokenTransferHandler(transferUCase, config)
	appRouter.POST("/token-transfer", transferHandler.Transfer)
	appRouter.GET("/token-transfer/histories", transferHandler.GetTokenTransferHistories)

	// SECTION: block state
	blockstateRepository := blockstate.NewBlockstateRepository(db)
	blockstateUcase := blockstate.NewBlockStateUCase(blockstateRepository)

	// SECTION: payment wallet
	paymentWalletRepository := paymentwallet.NewPaymentWalletRepository(db, config)

	// SECTION: payment order
	paymentOrderRepository := paymentorder.NewPaymentOrderRepository(db)
	paymentOrderUCase := paymentorder.NewPaymentOrderUCase(db, paymentOrderRepository, paymentWalletRepository, blockstateRepository, cacheRepository)
	paymentOrderHandler := paymentorder.NewPaymentOrderHandler(paymentOrderUCase, config)
	appRouter.POST("/payment-orders", paymentOrderHandler.CreateOrders)

	// SECTION: init payment wallets
	err := utils.InitPaymentWallets(ctx, config, paymentWalletRepository)
	if err != nil {
		log.LG.Fatalf("init payment wallets error: %v", err)
	}

	// Create a new payment order queue
	paymentOrderQueue, err := queue.NewQueue[dto.PaymentOrderDTO](ctx, constants.MinQueueLimit, paymentOrderUCase.GetPendingPaymentOrders)
	if err != nil {
		log.LG.Fatalf("create payment order queue error: %v", err)
	}

	// SECTION: start workers
	latestBlockWorker := worker.NewLatestBlockWorker(cacheRepository, blockstateUcase, ethClient)
	go latestBlockWorker.Start(ctx)

	expiredOrderCatchupWorker := worker.NewExpiredOrderCatchupWorker(
		config,
		paymentOrderUCase,
		cacheRepository,
		config.Blockchain.SmartContract.USDTContractAddress,
		ethClient,
	)
	go expiredOrderCatchupWorker.Start(ctx)

	releaseWalletWorker := worker.NewReleaseWalletWorker(paymentOrderUCase)
	go releaseWalletWorker.Start(ctx)

	// SECTION: start events listener
	// start base event listener for common listener logic
	baseEventListener := listener.NewBaseEventListener(
		ethClient,
		cacheRepository,
		blockstate.NewBlockstateRepository(db),
		&config.Blockchain.StartBlockListener,
	)
	// register token transfer event listener
	tokenTransferListener, err := listener.NewTokenTransferListener(
		ctx,
		config,
		baseEventListener,
		paymentOrderUCase,
		config.Blockchain.SmartContract.USDTContractAddress,
		paymentOrderQueue,
	)
	if err != nil {
		log.LG.Errorf("Failed to initialize tokenTransferListener: %v", err)
		return
	}
	tokenTransferListener.Register(ctx)
	// run event listeners
	go func() {
		if err := baseEventListener.RunListener(ctx); err != nil {
			log.LG.Errorf("Error running event listeners: %v", err)
		}
	}()
}
