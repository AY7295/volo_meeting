package handler

import (
	"time"
	"volo_meeting/consts"
	"volo_meeting/internal/hub"
	"volo_meeting/internal/usecase/meeting/service"
	"volo_meeting/lib/callback"
	error2 "volo_meeting/lib/error"

	"github.com/gin-gonic/gin"
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

// func JoinMeetingRoom(ctx *gin.Context) {
// 	id := ctx.Query("meeting_id")
// 	if len(id) != consts.DefaultMeetingIdSize {
// 		callback.Error(ctx, error2.InvalidMeetingId)
// 		return
// 	}

// 	bind := &hub.Device{JoinTime: time.Now().Unix()}
// 	if err := ctx.BindQuery(bind); err != nil {
// 		callback.Error(ctx, error2.New(consts.ParamError, err))
// 		return
// 	}

// 	callback.Error(ctx, service.JoinMeetingRoom(ctx, id, bind))
// }

func JoinMeetingRoom(ctx *gin.Context) {
	id := ctx.Query("meeting_id")
	bind := &hub.Device{
		Id:       ctx.Query("id"),
		Nickname: ctx.Query("nickname"),
		JoinTime: time.Now().Unix(),
	}

	callback.Error(ctx, service.JoinMeetingRoom(ctx, id, bind))
}
