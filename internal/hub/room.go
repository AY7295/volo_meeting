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
	member := &Member{
		Device: device,
		Room:   r,
		Conn:   conn,
	}
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
	Device *Device
	Room   *Room
	Conn   *ws.Conn
}

func (m *Member) setupEmitter() {
	m.Conn.On(consts.Message, func(data []byte) {
		message := &Message[jsoniter.RawMessage]{}
		err := jsoniter.Unmarshal(data, message)
		if err != nil {
			m.Conn.Emit(consts.Err, error2.New(consts.MarshalError, err))
			return
		}

		zap.L().Debug("receive message", zap.String("deviceId", m.Device.Id), zap.Any("event", message.Event), zap.Any("data", message.Data))

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
			m.Conn.Emit(consts.Err, error2.New(consts.ParamError, fmt.Errorf("unknown event type: %v", message.Event)))
		}
	})

	m.Conn.On(consts.Join, func() {
		zap.L().Debug("receive join", zap.String("deviceId", m.Device.Id))
		m.Conn.Send(&Message[[]*Device]{consts.Member, m.Room.getDevices(m.Device.Id)})

		broadcast(m.Room, &Message[[]*Device]{
			Event: consts.Member,
			Data: []*Device{
				m.Device,
			},
		}, m.Device.Id)
	})

	m.Conn.On(consts.Close, func() {
		zap.L().Debug("receive close", zap.String("deviceId", m.Device.Id))
		defer m.Conn.Close()
		m.Room.Members.Delete(m.Device.Id)
		broadcast(m.Room, &Message[DeviceId]{
			Event: consts.Leave,
			Data:  m.Device.Id,
		}, m.Device.Id)
	})

	m.Conn.On(consts.Err, func(err error) {
		zap.L().Debug("receive error", zap.Error(err), zap.String("deviceId", m.Device.Id))
		m.Conn.Send(&Message[error]{
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
		m.Conn.Send(&Message[error]{consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	zap.L().Debug("update info", zap.Any("newDevice", device), zap.Any("oldDevice", m.Device))
	m.Device.Nickname = device.Nickname
	broadcast(m.Room, &Message[*Device]{
		Event: consts.Device,
		Data:  m.Device,
	}, deviceId)
}

// forwarding descp: forward message to specific device by Data.Id
func (m *Member) forwarding(deviceId DeviceId, message *Message[jsoniter.RawMessage]) {
	data := make([]Data, 0, m.Room.Members.Len()-1)
	err := jsoniter.Unmarshal(message.Data, &data)
	if err != nil {
		zap.L().Error("unmarshal error", zap.Error(err))
		m.Conn.Send(&Message[error]{consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	zap.L().Debug("forwarding", zap.String("deviceId", deviceId), zap.Any("event", message.Event), zap.Any("forwarding data", data))

	deliver(m.Room, &Message[[]Data]{
		Event: message.Event,
		Data:  data,
	}, deviceId)
}

// broadcast descp: broadcast message to all devices in room, except exceptions
func broadcast[T any](r *Room, message *Message[T], exceptions ...DeviceId) {
	fn := func(key DeviceId, value *Member) {
		value.Conn.Send(message)
	}

	r.Members.Range(fn, defaultExcept(exceptions...))
}

// deliver descp: deliver message to specific device by Data.Id, and change Data.Id to fromId
func deliver(r *Room, message *Message[[]Data], fromId DeviceId) {
	set := make(map[DeviceId]*Message[[]Data], len(message.Data))
	for i, data := range message.Data {
		set[data.Id] = &Message[[]Data]{
			Event: message.Event,
			Data:  []Data{message.Data[i]},
		}
	}

	fn := func(key DeviceId, value *Member) {
		msg, ok := set[key]
		if ok {
			for i := 0; i < len(msg.Data); i++ {
				(*msg).Data[i].Id = fromId
			}
			value.Conn.Send(msg)
		}
	}

	r.Members.Range(fn)
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
