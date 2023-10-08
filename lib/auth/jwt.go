package auth

import (
	"github.com/gin-gonic/gin"
)

func Token(ctx *gin.Context) {
	ctx.Set("uid", ctx.GetHeader("Authorization"))
	ctx.Next()
}
