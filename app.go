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
	Register(id string, w http.ResponseWriter, r *http.Request, recv Receiver, data UserData) (c ClientProxyer, err error)
	Unregister(id string)
	Count() int
	CountById() int
	SendTo(id string, b []byte)
	SendAll(b []byte)
}

type app struct {
	key string //app key
	*connpool
	receiver   Receiver
	register   chan *client
	unregister chan *client
}

type Sender interface {
	SendTo(id string, b []byte)
	SendAll(b []byte)
}

type Receiver interface {
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
		register:   make(chan *client),
		unregister: make(chan *client),
	}
	return
}

func (a *app) Register(id string, w http.ResponseWriter, r *http.Request, recv Receiver, data UserData) (c ClientProxyer, err error) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {

		return
	}
	client := newClient(id, ws, a, recv, data)
	a.Join(client)
	c = client
	return

}

func (a *app) Unregister(id string) {
	a.RemoveById(id)
	return
}
