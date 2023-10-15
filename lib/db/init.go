package db

import (
	"volo_meeting/lib/db/mysql"
	"volo_meeting/lib/db/redis"
)

func Init() {
	mysql.Init()
	redis.Init()
}
