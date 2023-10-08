package model

import (
	"gorm.io/gorm"
	"volo_meeting/lib/db/mysql"
)

var (
	instance *gorm.DB
)

func Instance() *gorm.DB {
	return instance
}

func Init() {
	instance = mysql.Instance()
	err := instance.AutoMigrate(
		&Meeting{},
		&Device{},
	)
	if err != nil {
		panic(err)
	}
}
