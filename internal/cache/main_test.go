package cache

import (
	"os"
	"testing"
	"volo_meeting/config"
	"volo_meeting/lib/db/redis"
)

func TestMain(m *testing.M) {
	config.Init()
	redis.Init()
	Init()

	m.Run()

	os.Exit(0)
}
