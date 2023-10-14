package ws

import (
	"time"
	"volo_meeting/consts"
	error2 "volo_meeting/lib/error"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Conn struct {
	DeviceId string
	*emission.Emitter
	socket *websocket.Conn
	timer  *time.Timer
	closed chan struct{}
}

func NewConn(socket *websocket.Conn, deviceId string) *Conn {
	conn := &Conn{
		DeviceId: deviceId,
		Emitter:  emission.NewEmitter(),
		socket:   socket,
		closed:   make(chan struct{}),
	}

	conn.socket.SetCloseHandler(func(code int, text string) error {
		zap.L().Info("websocket Close", zap.Int("code", code), zap.String("text", text))
		conn.Emit(consts.Close)
		return nil
	})

	conn.RecoverWith(func(i1, i2 interface{}, err error) {
		zap.L().Error("emitter panic", zap.String("deviceId", deviceId), zap.Error(err), zap.Any("i1", i1), zap.Any("i2", i2))
	})

	return conn
}

func (conn *Conn) Listen() {
	conn.timer = time.AfterFunc(consts.KeepaliveInterval, func() {
		conn.Emit(consts.Close)
	})

	for {
		select {
		case <-conn.closed:
			return
		default:
			message, err := conn.receive()
			if err == nil {
				conn.Emit(consts.Message, message)
				continue
			}

			zap.L().Error("read message error", zap.Error(err))
			conn.Emit(consts.Close)
			return
		}
	}
}

func (conn *Conn) receive() ([]byte, error) {
	if err := conn.isClosed(); err != nil {
		return nil, err
	}

	_, message, err := conn.socket.ReadMessage()
	return message, err
}

func (conn *Conn) Send(data any) {
	if err := conn.isClosed(); err != nil {
		zap.L().Info("websocket is closed", zap.Any("data", data))
		return
	}

	message, err := jsoniter.Marshal(data)
	if err != nil {
		conn.Emit(consts.Err, err)
		return
	}

	err = conn.socket.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		zap.L().Error("websocket write message error", zap.Error(err))
		conn.Emit(consts.Close)
	}
}

func (conn *Conn) KeepAlive() {
	conn.timer.Reset(consts.KeepaliveInterval)
}

func (conn *Conn) isClosed() error {
	select {
	case <-conn.closed:
		return error2.InvalidClosedSocket
	default:
		return nil
	}
}

func (conn *Conn) Close() {
	close(conn.closed)
	err := conn.socket.Close()
	if err != nil {
		zap.L().Error("Close ws conn error", zap.Error(err))
	}
}
