package handler

import (
	"errors"
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

func JoinMeetingRoom(ctx *gin.Context) {
	meetingId := ctx.Query("meeting_id")
	deviceId := ctx.Query("id")
	nickname := ctx.Query("nickname")
	if len(meetingId) == 0 || len(deviceId) == 0 || len(nickname) == 0 {
		callback.Error(ctx, error2.New(consts.ParamError, errors.New("empty params")))
		return
	}

	service.JoinMeetingRoom(ctx, meetingId, &hub.Device{
		Id:       deviceId,
		Nickname: nickname,
		JoinTime: time.Now().Unix(),
	})
}
