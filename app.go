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
	Sender
	SetRegisterHandler(f func(id string, s Sender))
	SetUnregisterHandler(f func(id string, s Sender))
}

type app struct {
	key string //app key
	*connpool
	receiver Receiver
}

type Sender interface {
	SendTo(id string, b []byte)
	SendAll(b []byte)
	SendToByRegex(id string, regex string, b []byte)
}

type Receiver interface {
	Receive(id string, s Sender, b []byte, data UserData)
}

func newApp(key string) (a *app) {

	cp := &connpool{
		lock:              new(sync.RWMutex),
		pool:              make(map[string]map[*client]UserData),
		registerHandler:   nil,
		unregisterHandler: nil,
	}
	a = &app{
		key:      key,
		connpool: cp,
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

func (a *app) SetRegisterHandler(f func(id string, s Sender)) {
	a.registerHandler = f

}
func (a *app) SetUnregisterHandler(f func(id string, s Sender)) {
	a.unregisterHandler = f

}
