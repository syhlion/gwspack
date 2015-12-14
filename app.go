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
	Register(id string, w http.ResponseWriter, r *http.Request, data UserData) (c ClientProxyer, err error)
	Unregister(id string)
	Count() int
	CountById() int
	SendTo(id string, b []byte)
	SendAll(b []byte)
}

type app struct {
	key                 string //app key
	clientSets          int    //多少連線數要多開一個 process
	connections         map[string]map[*client]UserData
	receiverProcessPool []chan<- int
	receiver            Receiver
	boradcast           chan []byte
	receive             chan message
	register            chan *client
	unregister          chan *client
	lock                *sync.RWMutex
}

type Sender interface {
	SendTo(id string, b []byte)
	SendAll(b []byte)
}

type Receiver interface {
	Receive(id string, s Sender, b []byte, data UserData)
}

func newApp(key string, r Receiver, clientSets int) (a *app) {

	a = &app{
		key:         key,
		connections: make(map[string]map[*client]UserData),
		receiver:    r,
		boradcast:   make(chan []byte),
		register:    make(chan *client),
		unregister:  make(chan *client),
		receive:     make(chan message),
		clientSets:  clientSets,
		lock:        new(sync.RWMutex),
	}
	return
}

func (a *app) Register(id string, w http.ResponseWriter, r *http.Request, data UserData) (c ClientProxyer, err error) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {

		return
	}
	c = newClient(id, ws, a, data)
	return

}

func (a *app) Unregister(id string) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	for c := range a.connections[id] {
		a.unregister <- c
	}
}

func (a *app) SendTo(id string, b []byte) {

	a.lock.RLock()
	defer a.lock.RUnlock()
	for c := range a.connections[id] {
		c.send <- b
	}

}

func (a *app) Count() int {
	var i int
	a.lock.RLock()
	defer a.lock.RUnlock()
	for k, _ := range a.connections {
		for _, _ = range a.connections[k] {
			i++
		}
	}
	return i
}

func (a *app) CountById() int {
	return len(a.connections)
}

func (a *app) SendAll(b []byte) {
	a.boradcast <- b
}

func (a *app) List() {
}

func (a *app) receiveHandle() chan<- int {
	end := make(chan int)
	go func() {
		for {
			select {
			case m := <-a.receive:
				a.receiver.Receive(m.clientId, a, m.content, m.data)
			case <-end:
				break
			}
		}
	}()
	return end
}
func (a *app) Run() {
	for {
		select {
		case c := <-a.register:
			a.lock.RLock()
			if v, ok := a.connections[c.id]; !ok {
				a.lock.RUnlock()
				m := make(map[*client]UserData)
				m[c] = c.data
				a.lock.Lock()
				a.connections[c.id] = m
				a.lock.Unlock()
			} else {
				a.lock.RUnlock()
				a.lock.Lock()
				v[c] = c.data
				a.lock.Unlock()
			}
			if IsExpand(a.Count(), a.clientSets, len(a.receiverProcessPool)) {
				c := a.receiveHandle()
				a.receiverProcessPool = append(a.receiverProcessPool, c)
			}

		case client := <-a.unregister:
			a.lock.RLock()
			if v, ok := a.connections[client.id]; ok {
				for c := range v {
					if c == client {
						a.lock.RUnlock()
						a.lock.Lock()
						delete(a.connections[client.id], client)
						close(client.send)
						a.lock.Unlock()
					}
				}
			} else {
				a.lock.RUnlock()
			}

			if IsReduce(a.Count(), a.clientSets, len(a.receiverProcessPool)) {

				a.receiverProcessPool = append(a.receiverProcessPool[:0], a.receiverProcessPool[1:]...)
			}
		case message := <-a.boradcast:
			for _, clientMap := range a.connections {
				for client := range clientMap {
					client.send <- message
				}
			}
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
