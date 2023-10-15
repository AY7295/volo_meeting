package ws

import (
	"time"
	"volo_meeting/consts"
	error2 "volo_meeting/lib/error"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Conn struct {
	*emission.Emitter

	socket *websocket.Conn
	timer  *time.Timer
	closed chan struct{}
}

func NewConn(socket *websocket.Conn) *Conn {
	conn := &Conn{
		Emitter: emission.NewEmitter(),
		socket:  socket,
		closed:  make(chan struct{}),
	}

	conn.RecoverWith(func(i1, i2 interface{}, err error) {
		zap.L().Error("emitter panic", zap.Error(err), zap.Any("code", i1), zap.Any("error-msg", i2))
	})

	return conn
}

func (conn *Conn) Listen() {
	// descp keepalive
	//conn.timer = time.AfterFunc(consts.KeepaliveInterval, func() {
	//	conn.Emit(consts.Close)
	//})

	for {
		select {
		case <-conn.closed:
			return
		default:
			_, message, err := conn.socket.ReadMessage()
			if err == nil {
				conn.Emit(consts.Message, message)
				continue
			}

			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				zap.L().Error("read message error", zap.Error(err))
			}

			conn.Emit(consts.Close)
			return
		}
	}
}

func (conn *Conn) Send(data any) {
	if err := conn.isClosed(); err != nil {
		zap.L().Info("send to closed conn", zap.Any("data", data))
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
