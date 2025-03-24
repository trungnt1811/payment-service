package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/database/types"
)

// postgreSQL implements SQLDBConnection interface
type postgreSQL struct {
	config *conf.DatabaseConfiguration
}

// NewPostgreSQLClient creates a new sql client instance
func NewPostgreSQLClient(config *conf.DatabaseConfiguration) types.SQLClient {
	return &postgreSQL{
		config: config,
	}
}

func (pgsql *postgreSQL) getDBConnectionURL() string {
	config := pgsql.config

	// Determine SSL mode based on the configuration
	sslMode := "disable"
	if config.SSLMode {
		sslMode = "enable"
	}

	// Format for PostgreSQL connection URL
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.DBHost, config.DBPort,
		config.DBUser, config.DBName, config.DBPassword, sslMode)
}

func (pgsql *postgreSQL) Connect() *gorm.DB {
	return pgsql.ConnectWithLogger(logger.Silent)
}

func (pgsql *postgreSQL) ConnectWithLogger(logLevel logger.LogLevel) *gorm.DB {
	// Get the PostgreSQL connection URL
	dbURL := pgsql.getDBConnectionURL()
	var db *gorm.DB
	var err error

	// Open the database connection with a custom log level
	db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		panic(err)
	}

	// Get the SQL database object
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Configure the connection pool
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum amount of time a connection may be reused

	return db
}
