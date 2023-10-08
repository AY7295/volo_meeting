package api

import (
	ginZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
	"volo_meeting/api/dev"
	"volo_meeting/api/meeting"
	"volo_meeting/lib/auth"
)

func Init() *gin.Engine {
	e := gin.New()
	e.Use(
		ginZap.Ginzap(zap.L(), time.RFC3339, false),
		ginZap.RecoveryWithZap(zap.L(), viper.GetBool("DEBUG")),
	)

	api := e.Group("api")
	{
		v1 := api.Group("v1", auth.Debug)
		{
			meeting.InitApi(v1.Group("meeting"))
		}
	}

	dev.InitApi(api.Group("dev", auth.Debug))

	return e
}
