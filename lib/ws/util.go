package ws

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"volo_meeting/consts"
	error2 "volo_meeting/lib/error"
)

var upgrade = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	socket, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error("websocket upgrade error", zap.Error(err))
	}

	return socket, error2.New(consts.WSError, err)
}
