package types

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLClient interface {
	Connect() *gorm.DB
	ConnectWithLogger(logLevel logger.LogLevel) *gorm.DB
}
