package handler

import (
	"github.com/gin-gonic/gin"
	"volo_meeting/consts"
	"volo_meeting/internal/usecase/meeting/service"
	"volo_meeting/lib/callback"
	error2 "volo_meeting/lib/error"
)

func AddMeeting(ctx *gin.Context) {
	callback.Final(ctx, func() (any, error) {
		return service.NewMeeting()
	})
}

func GetMemberList(ctx *gin.Context) {
	id := ctx.Query("id")
	if len(id) != consts.DefaultMeetingIdSize {
		callback.Error(ctx, error2.InvalidMeetingId)
		return
	}

	callback.Final(ctx, func() (any, error) {
		return service.GetMemberList(id)
	})
}
