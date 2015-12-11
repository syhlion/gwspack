package gwspack

import (
	"github.com/gorilla/websocket"
	"net/http"
)

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
	clientSets          int    //多少連線數要多開一個 process
	connections         map[string]map[*client]bool
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
	Receive(id string, s Sender, b []byte)
}

func NewApp(key string, r Receiver, clientSets int) (app *App) {

	app = &App{
		key:         key,
		connections: make(map[string]map[*client]bool),
		receiver:    r,
		boradcast:   make(chan []byte),
		register:    make(chan *client),
		unregister:  make(chan *client),
		receive:     make(chan Message),
		clientSets:  clientSets,
	}
	return
}

func (a *App) Register(id string, w http.ResponseWriter, r *http.Request) (c ClientProxyer, err error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c = newClient(id, ws, a)
	return

}

func (a *App) Unregister(tag string) {
	for c := range a.connections[tag] {
		a.unregister <- c
	}
}

func (a *App) SendTo(tag string, b []byte) {

	for c := range a.connections[tag] {
		c.send <- b
	}

}

func (a *App) Count() int {
	var i int
	for k, _ := range a.connections {
		for _, _ = range a.connections[k] {
			i++
		}
	}
	return i
}

func (a *App) CountByTag() int {
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
		case c := <-a.register:
			if v, ok := a.connections[c.id]; !ok {
				m := make(map[*client]bool)
				m[c] = true
				a.connections[c.id] = m
			} else {
				v[c] = true
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
