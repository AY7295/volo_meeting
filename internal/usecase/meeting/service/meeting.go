package service

import (
	"context"
	"errors"
	"volo_meeting/consts"
	"volo_meeting/internal/cache"
	"volo_meeting/internal/hub"
	"volo_meeting/internal/model"
	"volo_meeting/internal/usecase/meeting/request"
	"volo_meeting/lib/callback"
	error2 "volo_meeting/lib/error"
	"volo_meeting/lib/id"
	"volo_meeting/lib/ws"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewMeeting() (*request.MeetingInfo, error) {
	mMeeting, err := createMeeting()
	if err != nil {
		return nil, err
	}

	mMeeting.FriendlyId, err = generateFriendlyId(mMeeting.Id)
	if err != nil {
		return nil, err
	}

	return &request.MeetingInfo{
		Id:         mMeeting.Id,
		FriendlyId: mMeeting.FriendlyId,
	}, nil
}

func GetMemberList(id string) ([]*request.Device, error) {

	all, err := cache.ZRevRange(context.Background(), id, 0, -1)
	if err != nil {
		zap.L().Error("get member list error", zap.Error(err))
	}

	devices := make([]*request.Device, 0, len(all))
	for _, v := range all {
		device := &request.Device{}
		err = device.UnmarshalBinary([]byte(v))
		if err != nil {
			zap.L().Error("unmarshal device error", zap.Error(err))
			continue
		}
		devices = append(devices, device)
	}

	return devices, error2.New(consts.CacheError, err)
}

func JoinMeetingRoom(ctx *gin.Context, id string, device *hub.Device) {
	// descp get meeting id from cache
	// if len(id) != consts.DefaultMeetingIdSize {
	// 	if id, err = cache.Get(ctx, id); err != nil {
	// 		return err
	// 	}
	// }

	err := checkEndedMeeting(id)
	if err != nil {
		callback.Error(ctx, err)
		return
	}
	err = appendDevice(id, device.Id)
	if err != nil {
		callback.Error(ctx, err)
		return
	}

	socket, err := ws.Upgrade(ctx.Writer, ctx.Request)
	if err != nil {
		callback.Error(ctx, err)
		return
	}
	conn := ws.NewConn(socket)

	go conn.Listen()

	hub.Global.JoinRoom(id, device, conn)
}

// createMeeting create a meeting and retry 3 times if failed
func createMeeting() (*model.Meeting, error) {
	var err error
	mMeeting := &model.Meeting{}
	for i := 0; i < 3; i++ {
		mMeeting.Id, err = id.GetMeetingId()
		err = mMeeting.Create(model.Instance())
		if err == nil {
			return mMeeting, nil
		}
		zap.L().Error("create meeting error", zap.Error(err))
	}

	return nil, error2.New(consts.SeverError, err)
}

// generateFriendlyId generate a friendly id for meeting and redis.set nx 30 days
func generateFriendlyId(meetingId string) (string, error) {
	var (
		ok         bool
		err        error
		friendlyId string
	)
	for i := 0; i < 3; i++ {
		friendlyId, err = id.GetFriendlyId()
		if err != nil {
			err = error2.New(consts.SeverError, err)
			zap.L().Error("generate friendly id error", zap.Error(err))
			continue
		}
		ok, err = cache.SetNX(context.TODO(), friendlyId, meetingId, consts.FriendlyIdExpire)
		if !ok {
			continue
		}
		if err != nil {
			err = error2.New(consts.CacheError, err)
			zap.L().Error("set friendly id error", zap.Error(err))
			continue
		}

		break
	}

	return friendlyId, err
}

func appendDevice(meetingId, deviceId string) error {
	device := &model.Device{Id: deviceId}
	err := model.Instance().FirstOrCreate(device).Error
	if err != nil {
		zap.L().Error("create device error", zap.Error(err))
		return error2.New(consts.SqlError, err)
	}

	err = model.Instance().Model(&model.Meeting{Id: meetingId}).Association("Devices").Append(device)
	if err != nil {
		zap.L().Error("append device error", zap.Error(err))
	}

	return error2.New(consts.SqlError, err)
}

func checkEndedMeeting(id string) error {
	meeting := &model.Meeting{Id: id}
	err := meeting.FindById(model.Instance())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return error2.NotFound("meeting not found, id: " + id)
		}
		zap.L().Error("get meeting error", zap.Error(err))
		return error2.New(consts.SqlError, err)
	}

	if meeting.EndTime != nil {
		return error2.EndedMeeting
	}

	return nil
}
