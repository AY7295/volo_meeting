package meeting

import (
	"github.com/gin-gonic/gin"
	"volo_meeting/api/meeting/handler"
)

func InitApi(group *gin.RouterGroup) {
	group.GET("fast", handler.AddMeeting)
	group.GET("member", handler.GetMemberList)
}
