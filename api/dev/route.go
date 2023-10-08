package dev

import (
	"github.com/gin-gonic/gin"
)

func InitApi(group *gin.RouterGroup) {
	group.GET("ping", Pong)
}

func Pong(ctx *gin.Context) {

}
