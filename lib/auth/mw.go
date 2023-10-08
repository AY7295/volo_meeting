package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"volo_meeting/lib/callback"
	error2 "volo_meeting/lib/error"
)

func User(ctx *gin.Context) {
	uid, ok := ctx.Value("uid").(string)
	if !ok || uid == "" {
		callback.Error(ctx, error2.InvalidContext)
		return
	}

	// ctx.Set("user", user)
	ctx.Next()
}

func Debug(ctx *gin.Context) {
	if !viper.GetBool("DEBUG") {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Set("uid", ctx.GetHeader("Authorization"))
	ctx.Next()
}
