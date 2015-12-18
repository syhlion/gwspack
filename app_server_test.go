package gwspack

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

var testlock *sync.RWMutex = new(sync.RWMutex)

type testReceiver struct{ *testing.T }

func (t *testReceiver) Receive(tag string, s Sender, b []byte, data UserData) {
	s.SendAll(b)

}
func (t *testReceiver) AfterRegister(tag string, s Sender, data UserData) {
	return

}
func (t *testReceiver) AfterUnregister(tag string, s Sender, data UserData) {
	return

}

type testAfterRegister struct{ *testing.T }

func (t *testAfterRegister) Receive(tag string, s Sender, b []byte, data UserData) {
	s.SendAll(b)

}
func (t *testAfterRegister) AfterRegister(tag string, s Sender, data UserData) {
	s.SendAll([]byte("HIHI"))

}
func (t *testAfterRegister) AfterUnregister(tag string, s Sender, data UserData) {

}

func newWebScoetClient(a string) (wsConn *websocket.Conn) {
	u, err := url.Parse(a)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsHeaders := http.Header{
		"Origin":                   {"http://local"},
		"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame"},
	}
	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsConn, _, err = websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	return wsConn
}

func TestRegister(t *testing.T) {
	tag := [3]string{"a", "a", "b"}

	i := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		a := Get("testKey")
		testlock.RLock()
		tt := tag[i]
		testlock.RUnlock()
		c, err := a.Register(tt, w, r, &testReceiver{t}, nil)
		if err != nil {
			t.Error(err)
		}
		testlock.Lock()
		i++
		testlock.Unlock()
		c.Listen()
		return

	}))
	ws := newWebScoetClient(ts.URL)
	ws2 := newWebScoetClient(ts.URL)
	ws3 := newWebScoetClient(ts.URL)
	defer func() {
		ws.Close()
		ws2.Close()
		ws3.Close()
		ts.Close()
	}()
	ap := Get("testKey")
	if ap.Count() != 3 {
		t.Error("count error", ap.Count())
	}
	if ap.CountById() != 2 {
		t.Error("count by tag error", ap.Count())
	}
	ma := Info()
	if v, ok := ma["testKey"]; !ok && v != ap.Count() {
		t.Error("Info error", ma)
	}
	return

}
func TestSendAll(t *testing.T) {

	tag := [3]string{"a", "a", "b"}

	i := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		a := Get("testKey")
		testlock.RLock()
		tt := tag[i]
		testlock.RUnlock()
		c, err := a.Register(tt, w, r, &testReceiver{}, nil)
		if err != nil {
			t.Error(err)
		}
		testlock.Lock()
		i++
		testlock.Unlock()
		c.Listen()

	}))
	defer ts.Close()
	ws := newWebScoetClient(ts.URL)
	ws2 := newWebScoetClient(ts.URL)
	ws3 := newWebScoetClient(ts.URL)
	defer func() {
		ws.Close()
		ws2.Close()
		ws3.Close()
	}()

	wschan := make(chan int)
	ws2chan := make(chan int)
	ws3chan := make(chan int)
	go func() {
		_, message, err := ws.ReadMessage()
		if err == nil && string(message) == "aaa" {
			wschan <- 1
		}
		defer close(wschan)
	}()
	go func() {
		_, message, err := ws2.ReadMessage()
		if err == nil && string(message) == "aaa" {
			ws2chan <- 1
		}
		defer close(ws2chan)
	}()
	go func() {
		_, message, err := ws3.ReadMessage()
		if err == nil && string(message) == "aaa" {
			ws3chan <- 1
		}
		defer close(ws3chan)
	}()
	err := ws.WriteMessage(websocket.TextMessage, []byte("aaa"))
	if err != nil {
		t.Error(err)
	}
	wsint := <-wschan
	ws2int := <-ws2chan
	ws3int := <-ws3chan
	if wsint != 1 && ws2int != 1 && ws3int != 1 {
		t.Error(wsint, ws2int, ws3int)
	}
	return
}
