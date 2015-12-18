package gwspack

import (
	"github.com/gorilla/websocket"
	"regexp"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type message struct {
	to      string
	regex   *regexp.Regexp
	content []byte
}
type ClientProxyer interface {
	Listen()
}
type client struct {
	id      string
	ws      *websocket.Conn
	app     *app
	send    chan []byte
	data    map[string]interface{}
	handler ClientHandler
}

func newClient(id string, ws *websocket.Conn, app *app, h ClientHandler, data map[string]interface{}) *client {
	return &client{
		id:      id,
		ws:      ws,
		send:    make(chan []byte, 4096),
		app:     app,
		data:    data,
		handler: h,
	}
}

func (c *client) write(msgType int, msg []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(msgType, msg)
}

func (c *client) readPump() {
	defer func() {
		c.ws.Close()
		c.app.disconnect <- c
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		if c.handler != nil {
			c.handler.Receive(c.id, c.app, msg, c.data)
		}
	}

}

func (c *client) Listen() {
	go c.writePump()
	c.readPump()
}

func (c *client) writePump() {
	t := time.NewTicker(pingPeriod)
	defer func() {
		c.ws.Close()
		c.app.disconnect <- c
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
