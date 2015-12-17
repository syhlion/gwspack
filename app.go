package gwspack

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type UserData map[string]interface{}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ClientController interface {
	Register(id string, w http.ResponseWriter, r *http.Request, recv ClientHandler, data UserData) (c ClientProxyer, err error)
	Unregister(id string)
	Count() int
	CountById() int
	Sender
}

type app struct {
	key string //app key
	*connpool
	send       chan message
	connect    chan *client
	unregister chan string
	disconnect chan *client
}

type Sender interface {
	SendTo(id string, b []byte)
	SendAll(b []byte)
	SendByRegex(regex string, b []byte)
}

type ClientHandler interface {
	Receive(id string, s Sender, b []byte, data UserData)
}

func newApp(key string) (a *app) {

	cp := &connpool{
		lock: new(sync.RWMutex),
		pool: make(map[string]map[*client]UserData),
	}
	a = &app{
		key:        key,
		connpool:   cp,
		send:       make(chan message, 1000),
		connect:    make(chan *client, 1000),
		unregister: make(chan string, 1000),
		disconnect: make(chan *client, 1000),
	}
	return
}

func (a *app) Register(id string, w http.ResponseWriter, r *http.Request, h ClientHandler, data UserData) (c ClientProxyer, err error) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {

		return
	}
	client := newClient(id, ws, a, h, data)
	a.connect <- client
	println("test")
	c = client
	return

}

func (a *app) SendTo(id string, b []byte) {
	m := message{id, b}
	a.send <- m
}

func (a *app) SendAll(b []byte) {
	m := message{"", b}
	a.send <- m
}

func (a *app) Unregister(id string) {
	a.unregister <- id
	return
}

func (a *app) run() {
	for {
		select {
		case c := <-a.connect:
			a.join(c)
		case m := <-a.send:
			if m.to == "" {
				a.sendAll(m.content)
			} else {
				a.sendTo(m.to, m.content)
			}
		case c := <-a.disconnect:
			a.remove(c)
		case id := <-a.unregister:
			a.removeById(id)
		}
	}

}
