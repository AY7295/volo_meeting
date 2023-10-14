package hub

import (
	"fmt"
	"volo_meeting/consts"
	"volo_meeting/internal/model"
	error2 "volo_meeting/lib/error"
	"volo_meeting/lib/tsmap"
	"volo_meeting/lib/ws"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type DeviceId = string

type Message[T any] struct {
	Event consts.Event `json:"event"`
	Data  T            `json:"data"`
}

type Data struct {
	From    DeviceId `json:"from;omitempty"`
	Content any      `json:"content"` // description, candidate
}

type Member struct {
	Device *Device
	Room   *Room
	Conn   *ws.Conn
}

type Device struct {
	Id          DeviceId `json:"id" redis:"id" binding:"required"`
	Nickname    string   `json:"nickname" redis:"nickname" binding:"required" `
	Description any      `json:"description" redis:"description"`
	JoinTime    int64    `json:"-" redis:"join_time"`
}

func newRoom(meeting *model.Meeting) *Room {
	return &Room{
		Members: tsmap.New[DeviceId, *Member](),
		Meeting: meeting,
	}
}

type Room struct {
	Members tsmap.TSMap[DeviceId, *Member]
	Meeting *model.Meeting
}

func (r *Room) Join(device *Device, conn *ws.Conn) {
	r.Members.Set(device.Id, &Member{
		Device: device,
		Room:   r,
		Conn:   conn,
	})

	r.setupMessaging(conn)
}

func (r *Room) setupMessaging(conn *ws.Conn) {

	conn.On(consts.Message, func(data []byte) {
		message := &Message[*jsoniter.RawMessage]{}
		err := jsoniter.Unmarshal(data, message)
		if err != nil {
			conn.Emit(consts.Err, error2.New(consts.MarshalError, err))
			return
		}
		switch message.Event {
		case consts.Description, consts.Candidate:
			r.forwarding(conn.DeviceId, message)
		case consts.Device:
			r.updateInfo(conn.DeviceId, message)
		case consts.Join:
			r.join(conn.DeviceId, message)
		case consts.KeepAlive:
			conn.KeepAlive()
		case consts.Leave:
			conn.Emit(consts.Close)
		default:
			conn.Emit(consts.Err, error2.New(consts.ParamError, fmt.Errorf("unkown event type: %v", message.Event)))
		}
	})

	conn.On(consts.Close, func() {
		defer conn.Close()
		r.Members.Delete(conn.DeviceId)
		broadcast(r, &Message[DeviceId]{
			Event: consts.Leave,
			Data:  conn.DeviceId,
		}, conn.DeviceId)
	})

	conn.On(consts.Err, func(err error) {
		conn.Send(&Message[error]{
			Event: consts.Error,
			Data:  err,
		})
	})
}

func (r *Room) updateInfo(deviceId DeviceId, message *Message[*jsoniter.RawMessage]) {
	device := &Device{}
	err := jsoniter.Unmarshal(*message.Data, device)
	if err != nil {
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.MarshalError, err))
		return
	}
	if device.Id != deviceId {
		zap.L().Error("device id not match")
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.ParamError, fmt.Errorf("device id: %v not match FromId :%v", deviceId, device.Id)))
		return
	}

	member, ok := r.Members.Get(device.Id)
	if !ok {
		return
	}
	member.Device.Description = device.Description
	member.Device.Nickname = device.Nickname

	broadcast(r, message, deviceId)
}

func (r *Room) forwarding(deviceId DeviceId, message *Message[*jsoniter.RawMessage]) {
	data := &Data{}
	err := jsoniter.Unmarshal(*message.Data, data)
	if err != nil {
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.MarshalError, err))
		return
	}
	if data.From != deviceId {
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.ParamError, fmt.Errorf("device id: %v not match FromId :%v", deviceId, data.From)))
		return
	}

	broadcast(r, message, deviceId)
}

func (r *Room) join(deviceId DeviceId, message *Message[*jsoniter.RawMessage]) {
	data := &Data{}
	err := jsoniter.Unmarshal(*message.Data, data)
	if err != nil {
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.MarshalError, err))
		return
	}
	if data.From != deviceId {
		sendTo[error](r, deviceId, consts.Error, error2.New(consts.ParamError, fmt.Errorf("device id: %v not match FromId :%v", deviceId, data.From)))
		return
	}

	sendTo[[]*Device](r, deviceId, consts.Join, r.getDevices(deviceId))

	broadcast(r, message, deviceId)
}

func (r *Room) getDevices(exceptions ...DeviceId) []*Device {
	devices := make([]*Device, 0, r.Members.Len()-len(exceptions))
	fn := func(key DeviceId, value *Member) {
		devices = append(devices, value.Device)
	}

	r.Members.Range(fn, defaultExcept(exceptions...))

	return devices
}

func broadcast[T any](r *Room, message *Message[T], exceptions ...DeviceId) {
	fn := func(key DeviceId, value *Member) {
		value.Conn.Send(message)
	}

	r.Members.Range(fn, defaultExcept(exceptions...))
}

func sendTo[T any](r *Room, deviceId DeviceId, event consts.Event, data any) {
	member, ok := r.Members.Get(deviceId)
	if !ok {
		return
	}
	member.Conn.Send(&Message[T]{
		Event: event,
		Data:  data,
	})
}

func defaultExcept(exceptions ...DeviceId) func(key DeviceId, value *Member) bool {
	set := make(map[DeviceId]struct{}, len(exceptions))
	for _, e := range exceptions {
		set[e] = struct{}{}
	}
	return func(key DeviceId, value *Member) bool {
		_, ok := set[key]
		return ok
	}
}
