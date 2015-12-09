package wsexchange

import (
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type AppCollection struct {
	lock *sync.RWMutex
	apps map[string]*App
}

func NewAppCollection() *AppCollection {
	return &AppCollection{
		lock: new(sync.RWMutex),
		apps: make(map[string]*App),
	}
}

func (ac *AppCollection) Join(a *App) (app *App) {

	ac.lock.Lock()
	defer ac.lock.Unlock()
	if _, ok := ac.apps[a.key]; !ok {
		ac.apps[a.key] = a
	}
	app = ac.apps[a.key]
	return
}

func (ac *AppCollection) Get(key string) (app *App, err error) {
	ac.lock.RLock()
	defer ac.lock.RUnlock()
	if _, ok := ac.apps[key]; ok {
		app = ac.apps[key]
	} else {
		err = errors.New("empty")
	}
	return
}

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
		//暫時不實做推送到 送出頻道 目前是 readonly
		//c.Send <- msg

		c.app.receive <- Message{c.tag, msg}
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Message struct {
	Tag     string
	Content []byte
}

type App struct {
	key                 string //app key
	processCount        int    //多少連線數要多開一個 process
	connections         map[*client]bool
	receiverProcessPool []chan<- int
	receiver            Receiver
	boradcast           chan []byte
	receive             chan Message
	register            chan *client
	unregister          chan *client
}

type Sender interface {
	SendTo(tag string, b []byte)
	SendAll(b []byte)
}

type Receiver interface {
	Receive(tag string, s Sender, b []byte) error
}

func NewApp(key string, r Receiver) (app *App) {

	app = &App{
		key:         key,
		connections: make(map[*client]bool),
		receiver:    r,
		boradcast:   make(chan []byte),
		register:    make(chan *client),
		unregister:  make(chan *client),
	}
	return
}

func (a *App) Register(tag string, w http.ResponseWriter, r *http.Request) (c ClientProxyer, err error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c = newClient(tag, ws, a)
	return

}

func (a *App) Unregister(tag string) {
	for c := range a.connections {
		if c.tag == tag {
			a.unregister <- c
		}
	}
}

func (a *App) SendTo(tag string, b []byte) {

	for c := range a.connections {
		if c.tag == tag {
			c.send <- b
		}
	}

}

func (a *App) Count() int {
	return len(a.connections)
}

func (a *App) SendAll(b []byte) {
	a.boradcast <- b
}

func (a *App) receiveHandle() chan<- int {
	end := make(chan int)
	go func() {
		for {
			select {
			case m := <-a.receive:
				a.receiver.Receive(m.Tag, a, m.Content)
			case <-end:
				break
			}
		}
	}()
	return end
}
func (a *App) Run() {
	for {
		select {
		case client := <-a.register:
			a.connections[client] = true
			if a.isExpand() {
				c := a.receiveHandle()
				a.receiverProcessPool = append(a.receiverProcessPool, c)
			}

		case client := <-a.unregister:
			if _, ok := a.connections[client]; ok {
				delete(a.connections, client)
				close(client.send)
			}
			if !a.isReduce() {

				a.receiverProcessPool = append(a.receiverProcessPool[:0], a.receiverProcessPool[1:]...)
			}
		case message := <-a.boradcast:
			for client := range a.connections {
				client.send <- message
			}
		}

	}
	defer func() {
		close(a.boradcast)
		close(a.register)
		close(a.unregister)
	}()
}

func (a *App) isExpand() bool {
	count := a.Count() / a.processCount
	residue := a.Count() % a.processCount
	length := len(a.receiverProcessPool)
	if length == 0 || length < count || (count%residue) > (a.processCount/2) {
		return true
	}
	return false
}
func (a *App) isReduce() bool {
	count := a.Count() / a.processCount
	residue := a.Count() % a.processCount
	length := len(a.receiverProcessPool)
	if length > 1 {
		return false
	}
	if length > count && (count%residue) < (a.processCount/2) {
		return true
	}
	return false

}
