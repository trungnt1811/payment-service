package instances

import (
	"sync"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/database/postgres"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	dbOnce     sync.Once
	dbInstance *gorm.DB
)

// DBInstance provides a singleton instance of the PostgreSQL database connection.
func DBInstance() *gorm.DB {
	dbOnce.Do(func() {
		logger.GetLogger().Info("Initializing PostgreSQL database connection...")

		// Get the configuration
		config := conf.GetConfiguration()

		// Create a new PostgreSQL client
		pgsqlClient := postgres.NewPostgreSQLClient(&config.Database)

		// Connect and store the database instance
		dbInstance = pgsqlClient.Connect()

		logger.GetLogger().Info("PostgreSQL database connection established.")
	})
	return dbInstance
}
