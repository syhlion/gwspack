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
	key        string //app key
	pool       *connpool
	receiver   Receiver
	boradcast  chan []byte
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
		pool:       cp,
		boradcast:  make(chan []byte),
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
	c = newClient(id, ws, a, recv, data)
	return

}

func (a *app) Unregister(id string) {
	a.pool.removeById(id)
	return
}

func (a *app) SendTo(id string, b []byte) {
	a.pool.sendTo(id, b)
	return

}

func (a *app) Count() int {
	return a.pool.count()
}

func (a *app) CountById() int {
	return a.pool.countById()
}

func (a *app) SendAll(b []byte) {
	a.boradcast <- b
}

func (a *app) List() {
}

func (a *app) Run() {
	for {
		select {
		case c := <-a.register:
			a.pool.join(c)
		case client := <-a.unregister:
			a.pool.remove(client)

		case message := <-a.boradcast:
			a.pool.sendAll(message)
		}

	}
	defer func() {
		close(a.boradcast)
		close(a.register)
		close(a.unregister)
	}()
}

func IsExpand(clientCount int, clientSets int, processCount int) bool {
	sets := clientCount / clientSets
	residue := clientCount % clientSets
	if processCount == 0 {
		return true
	}
	if processCount <= sets && residue > (clientSets/2) {
		return true
	}
	return false
}
func IsReduce(clientCount int, clientSets int, processCount int) bool {
	sets := clientCount / clientSets
	residue := clientCount % clientSets
	if processCount <= 1 {
		return false
	}
	if processCount >= sets && residue < (clientSets/2) {
		return true
	}
	return false

}
