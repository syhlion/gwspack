package wsexchange

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type ClientProxyer interface {
	Listen()
}
type client struct {
	tag  string
	ws   *websocket.Conn
	app  *App
	send chan []byte
}

func newClient(tag string, ws *websocket.Conn, app *App) *client {
	return &client{
		tag:  tag,
		ws:   ws,
		send: make(chan []byte, 1024),
		app:  app,
	}
}

func (c *client) write(msgType int, msg []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(msgType, msg)
}

func (c *client) readPump() {
	defer func() {
		c.ws.Close()
		c.app.unregister <- c
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		c.app.receive <- Message{c.tag, msg}
		println(string(msg))
	}

}

func (c *client) Listen() {
	c.app.register <- c
	go c.writePump()
	c.readPump()
}

func (c *client) writePump() {
	t := time.NewTicker(pingPeriod)
	defer func() {
		c.ws.Close()
		c.app.unregister <- c
		t.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-t.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}

		}
	}

}
