package server

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type myConn struct {
	conn    *websocket.Conn
	lastUse time.Time
}

func (c *myConn) Write(bytes []byte) (n int, err error) {
	slog.Info("Attempting socket write.")
	c.lastUse = time.Now()
	err = c.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		slog.Error("bad socket write", "err", err.Error())
		return 0, err
	}
	slog.Info("Socket seemed to work.")
	return len(bytes), nil
}
