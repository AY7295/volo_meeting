package mysql

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
	"volo_meeting/consts"
	customLog "volo_meeting/lib/log"
)

var (
	db *gorm.DB
)

func Init() {
	var err error
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf(viper.GetString("mysql.volo.dsn"), viper.GetString("MYSQL_PASSWORD")),
	}), gormConfig(consts.DefaultSQLLogFilePath))
	if err != nil {
		panic(err)
	}
}

func Instance() *gorm.DB {
	return db
}

func gormConfig(logfile string) *gorm.Config {

	mode := logger.Error
	if viper.GetBool("DEBUG") {
		mode = logger.Info
	}

	level := zapcore.ErrorLevel
	if viper.GetBool("DEBUG") {
		level = zapcore.InfoLevel
	}

	return &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(
			customLog.CreateOsLogger(logfile, level),
			logger.Config{
				SlowThreshold: time.Millisecond,
				LogLevel:      mode,
				Colorful:      viper.GetBool("DEBUG"),
			},
		),
		PrepareStmt: true,
	}
}
