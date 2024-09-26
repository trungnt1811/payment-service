package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/genefriendway/onchain-handler/conf"
)

func GetDBConnectionURL() string {
	config := conf.GetConfiguration()
	// Format for PostgreSQL connection URL
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		config.Database.DbHost, config.Database.DbPort,
		config.Database.DbUser, config.Database.DbName, config.Database.DbPassword)
}

func DBConn() *gorm.DB {
	// Get the PostgreSQL connection URL
	dbUrl := GetDBConnectionURL()
	fmt.Println(dbUrl)
	var db *gorm.DB
	var err error

	// Open a connection using the PostgreSQL driver
	db, err = gorm.Open(postgres.Open(dbUrl), GetGormConfig())
	if err != nil {
		panic(err)
	}

	// Get the generic SQL database object
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

func DBConnWithLoglevel(logMode logger.LogLevel) *gorm.DB {
	// Get the PostgreSQL connection URL
	dbUrl := GetDBConnectionURL()
	var db *gorm.DB
	var err error

	// Open the database connection with a custom log level
	db, err = gorm.Open(postgres.Open(dbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Database CONNECTED")

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

func GetGormConfig() *gorm.Config {
	// Set default log mode to Info
	logMode := logger.Info
	return &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}
}
