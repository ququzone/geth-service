package web

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/ququzone/geth-service/service"
)

var (
	upgrader = websocket.Upgrader{}
)

type WebsocketPool struct {
	sync.RWMutex
	connections map[*websocket.Conn]bool
}

func (s *WebsocketPool) Receive(message string) error {
	s.RLock()

	brokens := make([]*websocket.Conn, 0)

	for conn := range s.connections {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Printf("push %s error: %v\n", conn.RemoteAddr(), err)
			brokens = append(brokens, conn)
		}
	}
	s.RUnlock()

	if len(brokens) > 0 {
		s.Lock()
		defer s.Unlock()
		for _, broken := range brokens {
			delete(s.connections, broken)
		}
	}
	return nil
}

var wsPool *WebsocketPool

func NewWebsocketPool() *WebsocketPool {
	wsPool = &WebsocketPool{
		connections: make(map[*websocket.Conn]bool),
	}
	return wsPool
}

func Websocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	wsPool.Lock()
	wsPool.connections[ws] = true
	wsPool.Unlock()

	defer func(conn *websocket.Conn) {
		wsPool.Lock()
		delete(wsPool.connections, conn)
		wsPool.Unlock()
	}(ws)

	hs, err := service.GetHeaderService()
	if err != nil {
		_ = ws.Close()
	}
	if err := ws.WriteMessage(websocket.TextMessage, []byte(hs.Json())); err != nil {
		_ = ws.Close()
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
