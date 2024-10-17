package internal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/conf/database"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/middleware"
	routeV1 "github.com/genefriendway/onchain-handler/internal/route"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

func RunApp(config *conf.Configuration) {
	// Create a context to handle shutdown signals
	ctx, cancel := context.WithCancel(context.Background())

	// Use release mode in production
	if config.Env == "PROD" {
		log.LG = log.NewZerologLogger(os.Stdout, zerolog.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.LG = log.NewZerologLogger(os.Stdout, zerolog.DebugLevel)
	}

	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLogger(log.LG.Instance))
	r.Use(gin.Recovery())

	db := database.DBConnWithLoglevel(logger.Info)

	// SECTION: init eth client
	ethClient, err := ethclient.Dial(config.Blockchain.RpcUrl)
	if err != nil {
		log.LG.Fatalf("failed to connect to eth client: %v", err)
	}
	defer ethClient.Close()

	// SECTION: init cache
	cacheClient := caching.NewGoCacheClient()
	cacheRepository := caching.NewCachingRepository(ctx, cacheClient)

	// SECTION: register routes with context
	routeV1.RegisterRoutes(ctx, r, config, db, cacheRepository, ethClient)

	// Register general handlers
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("%s is still alive", config.AppName),
		})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// SECTION: run Gin router
	go func() {
		if err := r.Run(fmt.Sprintf("0.0.0.0:%v", config.AppPort)); err != nil {
			log.LG.Fatalf("failed to run gin router: %v", err)
		}
	}()

	// Handle shutdown signals
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	log.LG.Info("Shutting down gracefully...")
	cancel() // Cancel the context to stop the event listener
}
