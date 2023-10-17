package hub

import (
	"fmt"
	"sync/atomic"
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
	Id    int32        `json:"id"`
	Event consts.Event `json:"event"`
	Data  T            `json:"data"`
}

type Data struct {
	Id      DeviceId             `json:"id"`
	Content *jsoniter.RawMessage `json:"content"` // description, candidate
}

type Device struct {
	Id       DeviceId `json:"id"`
	Nickname string   `json:"nickname"`
	JoinTime int64    `json:"-"`
}

type Room struct {
	Members tsmap.TSMap[DeviceId, *Member]
	Meeting *model.Meeting
}

func newRoom(meeting *model.Meeting) *Room {
	return &Room{
		Members: tsmap.New[DeviceId, *Member](),
		Meeting: meeting,
	}
}

func (r *Room) Join(device *Device, conn *ws.Conn) {
	member, ok := r.Members.Get(device.Id)
	if ok {
		member.Conn.Close()
		r.Members.Delete(device.Id)
	}

	member = newMember(device, conn, r)

	r.Members.Set(device.Id, member)

	member.setupEmitter()

	conn.Emit(consts.Join)
}

func (r *Room) getDevices(exceptions ...DeviceId) []*Device {
	devices := make([]*Device, 0, r.Members.Len()-len(exceptions))
	fn := func(key DeviceId, value *Member) {
		devices = append(devices, value.Device)
	}

	r.Members.Range(fn, defaultExcept(exceptions...))

	return devices
}

type Member struct {
	autoIncrId *atomic.Int32
	Device     *Device
	Room       *Room
	Conn       *ws.Conn
}

func newMember(device *Device, conn *ws.Conn, room *Room) *Member {
	return &Member{
		autoIncrId: &atomic.Int32{},
		Device:     device,
		Room:       room,
		Conn:       conn,
	}
}

func (m *Member) NextId() int32 {
	return m.autoIncrId.Add(1)
}

func (m *Member) setupEmitter() {
	m.Conn.On(consts.Message, func(data []byte) {
		message := &Message[jsoniter.RawMessage]{}
		err := jsoniter.Unmarshal(data, message)
		if err != nil {
			m.Conn.Emit(consts.Err, error2.New(consts.MarshalError, err), consts.WrongMessageModel)
			return
		}

		if message.Id <= 0 {
			m.Conn.Emit(consts.Err, error2.New(consts.ParamError, fmt.Errorf("message id must be greater than 0")), consts.InvalidId)
			return
		}

		zap.L().Debug("receive message", zap.String("deviceId", m.Device.Id), zap.Any("message", message))

		switch message.Event {
		case consts.Description, consts.Candidate:
			m.forwarding(m.Device.Id, message)
		case consts.Device:
			m.updateInfo(m.Device.Id, message)
		case consts.KeepAlive:
			m.Conn.KeepAlive()
		case consts.Leave:
			m.Conn.Emit(consts.Close)
		default:
			m.Conn.Emit(consts.Err, error2.New(consts.ParamError, fmt.Errorf("unknown event type: %v", message.Event)), message.Id)
		}
	})

	m.Conn.On(consts.Join, func() {
		zap.L().Debug("receive join", zap.String("deviceId", m.Device.Id))

		sendTo(m, &Message[[]*Device]{m.NextId(), consts.Member, m.Room.getDevices(m.Device.Id)})

		broadcast(m.Room, consts.Member, []*Device{m.Device}, m.Device.Id)
	})

	m.Conn.On(consts.Close, func() {
		zap.L().Debug("receive close", zap.String("deviceId", m.Device.Id))

		m.Room.Members.Delete(m.Device.Id)

		broadcast(m.Room, consts.Leave, m.Device.Id, m.Device.Id)

		m.Conn.Close()
	})

	m.Conn.On(consts.Err, func(err error, messageId int32) {
		zap.L().Debug("receive error", zap.Error(err), zap.String("deviceId", m.Device.Id))
		sendTo(m, &Message[error]{
			Id:    messageId,
			Event: consts.Error,
			Data:  err,
		})
	})

	zap.L().Debug("setup emitter finished", zap.String("deviceId", m.Device.Id))
}

func (m *Member) updateInfo(deviceId DeviceId, message *Message[jsoniter.RawMessage]) {
	device := &Device{Id: deviceId}
	err := jsoniter.Unmarshal(message.Data, device)
	if err != nil {
		zap.L().Error("unmarshal error", zap.Error(err))
		sendTo(m, &Message[error]{message.Id, consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	zap.L().Debug("update info", zap.Any("newDevice", device), zap.Any("oldDevice", m.Device))
	m.Device.Nickname = device.Nickname

	broadcast(m.Room, consts.Device, m.Device, deviceId)
}

// forwarding descp: forward message to specific device by Data.Id
func (m *Member) forwarding(deviceId DeviceId, message *Message[jsoniter.RawMessage]) {
	data := make([]Data, 0, m.Room.Members.Len()-1)
	err := jsoniter.Unmarshal(message.Data, &data)
	if err != nil {
		zap.L().Error("unmarshal error", zap.Error(err))
		sendTo(m, &Message[error]{message.Id, consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	zap.L().Debug("forwarding", zap.String("deviceId", deviceId), zap.Any("event", message.Event), zap.Any("forwarding data", data))

	deliver(m.Room, message.Event, data, deviceId)
}

// sendTo descp sendTo force the conn send Message type
func sendTo[T any](member *Member, message *Message[T]) {
	zap.L().Debug("send message", zap.String("deviceId", member.Device.Id), zap.Any("event", message.Event), zap.Any("data", message.Data))
	member.Conn.Send(message)
}

// broadcast descp: broadcast message to all devices in room, except exceptions
func broadcast[T any](r *Room, event consts.Event, data T, exceptions ...DeviceId) {
	fn := func(deviceId DeviceId, member *Member) {
		sendTo(member, &Message[T]{
			Id:    member.NextId(),
			Event: event,
			Data:  data,
		})
	}

	r.Members.Range(fn, defaultExcept(exceptions...))
}

// deliver descp: deliver message to specific device by Data.Id, and change Data.Id to fromId
// message
func deliver(r *Room, event consts.Event, data []Data, fromId DeviceId) {
	set := make(map[DeviceId]*Message[[]Data], len(data))
	for i, d := range data {
		set[d.Id] = &Message[[]Data]{
			Event: event,
			Data:  []Data{data[i]},
		}
	}

	fn := func(deviceId DeviceId, member *Member) {
		msg, ok := set[deviceId]
		if ok {
			for i := 0; i < len(msg.Data); i++ {
				msg.Id = member.NextId()
				msg.Data[i].Id = fromId
			}
			sendTo(member, msg)
		}
	}

	r.Members.Range(fn)
}

func defaultExcept(exceptions ...DeviceId) func(key DeviceId, value *Member) bool {
	set := make(map[DeviceId]struct{}, len(exceptions))
	for _, e := range exceptions {
		set[e] = struct{}{}
	}
	return func(deviceId DeviceId, member *Member) bool {
		_, ok := set[deviceId]
		return ok
	}
}
