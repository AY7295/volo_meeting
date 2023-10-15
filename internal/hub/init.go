package hub

import (
	"errors"
	"go.uber.org/zap"
	"volo_meeting/consts"
	"volo_meeting/internal/model"
	error2 "volo_meeting/lib/error"
	"volo_meeting/lib/tsmap"
	"volo_meeting/lib/ws"

	"gorm.io/gorm"
)

var (
	Global = newHub()
)

func newHub() *hub {
	return &hub{
		rooms: tsmap.New[MeetingId, *Room](),
	}
}

type MeetingId = string

type hub struct {
	rooms tsmap.TSMap[MeetingId, *Room]
}

func (h *hub) GetRoom(meetingId MeetingId) (*Room, error) {
	room, ok := h.rooms.Get(meetingId)
	if ok {
		return room, nil
	}

	// descp first device in meeting
	meeting := &model.Meeting{Id: meetingId}
	err := meeting.FindById(model.Instance())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, error2.NotFound("meeting not found")
		}
		return nil, error2.New(consts.SqlError, err)
	}

	if err = meeting.StartNow(model.Instance()); err != nil {
		return nil, error2.New(consts.SqlError, err)
	}

	room = newRoom(meeting)
	h.rooms.Set(meetingId, room)

	return room, nil
}

func (h *hub) JoinRoom(meetingId MeetingId, device *Device, conn *ws.Conn) {
	room, err := h.GetRoom(meetingId)
	if err != nil {
		zap.L().Error("get room error", zap.Error(err))
		conn.Send(&Message[error]{consts.Error, err})
		return
	}

	room.Join(device, conn)

	conn.Emit(consts.Join)
}

func (h *hub) RemoveRoom(meetingId MeetingId) error {
	room, err := h.GetRoom(meetingId)
	if err != nil {
		return err
	}
	defer func() {
		err = room.Meeting.EndNow(model.Instance())
		if err != nil {
			zap.L().Error("end meeting error", zap.Error(err))
		}
	}()

	room.Members.Range(func(key MeetingId, value *Member) {
		go value.Conn.Emit(consts.Close)
	})
	h.rooms.Delete(meetingId)

	return nil
}
