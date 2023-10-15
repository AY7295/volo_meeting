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

type Member struct {
	Device *Device
	Room   *Room
	Conn   *ws.Conn
}

type Device struct {
	Id       DeviceId `json:"id" redis:"id"`
	Nickname string   `json:"nickname" redis:"nickname"`
	//Description any      `json:"description" redis:"description"`
	JoinTime int64 `json:"-" redis:"join_time"`
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
	r.Members.Set(device.Id, &Member{
		Device: device,
		Room:   r,
		Conn:   conn,
	})

	r.setupEmitter(conn)
}

func (r *Room) setupEmitter(conn *ws.Conn) {
	conn.On(consts.Message, func(data []byte) {
		message := &Message[jsoniter.RawMessage]{}
		err := jsoniter.Unmarshal(data, message)
		if err != nil {
			conn.Emit(consts.Err, error2.New(consts.MarshalError, err))
			return
		}

		zap.L().Debug("receive message", zap.String("deviceId", conn.DeviceId), zap.Any("event", message.Event), zap.Any("data", message.Data))

		switch message.Event {
		case consts.Description, consts.Candidate:
			r.forwarding(conn.DeviceId, message)
		case consts.Device:
			r.updateInfo(conn.DeviceId, message)
		case consts.KeepAlive:
			conn.KeepAlive()
		case consts.Leave:
			conn.Emit(consts.Close)
		default:
			conn.Emit(consts.Err, error2.New(consts.ParamError, fmt.Errorf("unknown event type: %v", message.Event)))
		}
	})

	conn.On(consts.Join, func() {
		zap.L().Debug("receive join", zap.String("deviceId", conn.DeviceId))
		r.join(conn.DeviceId)
	})

	conn.On(consts.Close, func() {
		zap.L().Debug("receive close", zap.String("deviceId", conn.DeviceId))
		defer conn.Close()
		r.Members.Delete(conn.DeviceId)
		broadcast(r, &Message[DeviceId]{
			Event: consts.Leave,
			Data:  conn.DeviceId,
		}, conn.DeviceId)
	})

	conn.On(consts.Err, func(err error) {
		zap.L().Debug("receive error", zap.Error(err), zap.String("deviceId", conn.DeviceId))
		conn.Send(&Message[error]{
			Event: consts.Error,
			Data:  err,
		})
	})

	zap.L().Debug("setup emitter finished", zap.String("deviceId", conn.DeviceId))
}

func (r *Room) updateInfo(deviceId DeviceId, message *Message[jsoniter.RawMessage]) {
	device := &Device{Id: deviceId}
	err := jsoniter.Unmarshal(message.Data, device)
	if err != nil {
		zap.L().Error("unmarshal error", zap.Error(err))
		sendTo[error](r, deviceId, &Message[error]{consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	member, ok := r.Members.Get(deviceId)
	if !ok {
		zap.L().Error("member not found in room", zap.String("deviceId", deviceId), zap.Any("meeting", r.Meeting))
		return
	}
	zap.L().Debug("update info", zap.Any("new device", device), zap.Any("old device", member.Device))
	member.Device.Nickname = device.Nickname

	zap.L().Debug("room dsevices", zap.String("meetingId", r.Meeting.Id), zap.Any("devices", r.getDevices()))

	broadcast(r, &Message[*Device]{
		Event: consts.Device,
		Data:  member.Device,
	}, deviceId)
}

// forwarding descp: forward message to specific device by Data.Id
func (r *Room) forwarding(deviceId DeviceId, message *Message[jsoniter.RawMessage]) {
	data := make([]Data, 0, r.Members.Len()-1)
	err := jsoniter.Unmarshal(message.Data, &data)
	if err != nil {
		zap.L().Error("unmarshal error", zap.Error(err))
		sendTo[error](r, deviceId, &Message[error]{consts.Error, error2.New(consts.MarshalError, err)})
		return
	}

	zap.L().Debug("forwarding", zap.String("deviceId", deviceId), zap.Any("event", message.Event), zap.Any("forwarding data", data))

	deliver(r, &Message[[]Data]{
		Event: message.Event,
		Data:  data,
	}, deviceId)
}

// join descp: send all devices in room to new device, and send new device to all devices in room
func (r *Room) join(deviceId DeviceId) {
	member, ok := r.Members.Get(deviceId)
	if !ok {
		zap.L().Error("member not found in room", zap.String("deviceId", deviceId), zap.Any("meeting", r.Meeting))
		return
	}

	sendTo(r, deviceId, &Message[[]*Device]{consts.Member, r.getDevices(deviceId)})
	broadcast(r, &Message[[]*Device]{
		Event: consts.Member,
		Data: []*Device{
			member.Device,
		},
	}, deviceId)
}

func (r *Room) getDevices(exceptions ...DeviceId) []*Device {
	devices := make([]*Device, 0, r.Members.Len()-len(exceptions))
	fn := func(key DeviceId, value *Member) {
		devices = append(devices, value.Device)
	}

	r.Members.Range(fn, defaultExcept(exceptions...))

	return devices
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
			value.Conn.Send(message)
		}
	}

	r.Members.Range(fn)
}

func sendTo[T any](r *Room, to DeviceId, message *Message[T]) {
	member, ok := r.Members.Get(to)
	if !ok {
		return
	}
	member.Conn.Send(message)
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
