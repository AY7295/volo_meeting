package ws

import (
	"fmt"
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

	conn.RecoverWith(func(event, listener interface{}, err error) {
		zap.L().Error("emitter panic", zap.Error(err), zap.Any("event", event), zap.String("listener", fmt.Sprintf("%v", listener)))
	})

	return conn
}

func (conn *Conn) Listen() {
	// descp keepalive : consts.KeepaliveInterval don't reset timer will close conn
	conn.timer = time.AfterFunc(consts.KeepaliveInterval, func() {
		conn.Emit(consts.Close)
	})

	for {
		select {
		case <-conn.closed:
			return
		case <-time.NewTicker(consts.KeepaliveInterval / 2).C: // keepAlive every 10 seconds
			conn.ping()
		default:
			msgType, message, err := conn.socket.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					zap.L().Error("read message error", zap.Error(err))
				}

				conn.Emit(consts.Close)
				return
			}

			conn.handleMessage(msgType, message)
		}
	}
}

func (conn *Conn) handleMessage(messageType int, data []byte) {
	if messageType == websocket.PongMessage {
		conn.keepAlive()
		return
	}

	conn.Emit(consts.Message, data)
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

func (conn *Conn) keepAlive() {
	conn.timer.Reset(consts.KeepaliveInterval)
}

func (conn *Conn) ping() {
	err := conn.socket.WriteMessage(websocket.PingMessage, []byte("k"))
	if err != nil {
		zap.L().Error("websocket write message error", zap.Error(err))
		conn.Emit(consts.Close)
	}
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
