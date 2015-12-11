package gwspack

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ClientController interface {
	Register(id string, w http.ResponseWriter, r *http.Request, data map[string]interface{}) (c ClientProxyer, err error)
	Unregister(id string)
	Count() int
	CountById() int
}

type app struct {
	key                 string //app key
	clientSets          int    //多少連線數要多開一個 process
	connections         map[string]map[*client]map[string]interface{}
	receiverProcessPool []chan<- int
	receiver            Receiver
	boradcast           chan []byte
	receive             chan message
	register            chan *client
	unregister          chan *client
}

type Sender interface {
	SendTo(id string, b []byte)
	SendAll(b []byte)
}

type Receiver interface {
	Receive(id string, s Sender, b []byte, data map[string]interface{})
}

func newApp(key string, r Receiver, clientSets int) (a *app) {

	a = &app{
		key:         key,
		connections: make(map[string]map[*client]map[string]interface{}),
		receiver:    r,
		boradcast:   make(chan []byte),
		register:    make(chan *client),
		unregister:  make(chan *client),
		receive:     make(chan message),
		clientSets:  clientSets,
	}
	return
}

func (a *app) Register(id string, w http.ResponseWriter, r *http.Request, data map[string]interface{}) (c ClientProxyer, err error) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {

		return
	}
	c = newClient(id, ws, a, data)
	return

}

func (a *app) Unregister(id string) {
	for c := range a.connections[id] {
		a.unregister <- c
	}
}

func (a *app) SendTo(id string, b []byte) {

	for c := range a.connections[id] {
		c.send <- b
	}

}

func (a *app) Count() int {
	var i int
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
			if v, ok := a.connections[c.id]; !ok {
				m := make(map[*client]map[string]interface{})
				m[c] = c.data
				a.connections[c.id] = m
			} else {
				v[c] = c.data
			}
			if IsExpand(a.Count(), a.clientSets, len(a.receiverProcessPool)) {
				c := a.receiveHandle()
				a.receiverProcessPool = append(a.receiverProcessPool, c)
			}

		case client := <-a.unregister:
			for c := range a.connections[client.id] {
				if c == client {
					delete(a.connections[client.id], client)
					close(client.send)
				}
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
