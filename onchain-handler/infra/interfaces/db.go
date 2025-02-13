package interfaces

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLDBConnection interface {
	Connect() *gorm.DB
	ConnectWithLogger(logLevel logger.LogLevel) *gorm.DB
}
