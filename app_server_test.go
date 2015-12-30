package gwspack

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

var testlock *sync.RWMutex = new(sync.RWMutex)

type testReceiver struct{ *testing.T }

func (t *testReceiver) Receive(s Sender, b []byte) {
	s.SendAll(b)

}
func (t *testReceiver) GetUserData() UserData {
	return nil
}

type testAfterRegister struct{ *testing.T }

func (t *testAfterRegister) Receive(s Sender, b []byte) {
	s.SendAll(b)

}

func newWebScoetClient(a string) (wsConn *websocket.Conn, err error) {
	u, err := url.Parse(a)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsHeaders := http.Header{
		"Origin": {a},
	}
	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsConn, _, err = websocket.NewClient(rawConn, u, wsHeaders, 4096, 4096)
	if err != nil {
		fmt.Println(u.Host)
		fmt.Println(err)
		return
	}
	return
}

func TestRegister(t *testing.T) {
	tag := [5]string{"a", "a", "b"}

	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := Get("testKey")
		testlock.RLock()
		tt := tag[i]
		testlock.RUnlock()
		c, err := a.Register(tt, w, r, &testReceiver{t})
		if err != nil {
			t.Error(err)
			return
		}
		testlock.Lock()
		i++
		testlock.Unlock()
		c.Listen()
		return

	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	url := "ws://" + u.Host
	ws, err := newWebScoetClient(url)
	ws2, err := newWebScoetClient(url)
	ws3, err := newWebScoetClient(url)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		ws.Close()
		ws2.Close()
		ws3.Close()
	}()
	ap := Get("testKey")
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
		c, err := a.Register(tt, w, r, &testReceiver{})
		if err != nil {
			t.Error(err)
		}
		testlock.Lock()
		i++
		testlock.Unlock()
		c.Listen()

	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	url := "ws://" + u.Host
	ws, err := newWebScoetClient(url)
	ws2, err := newWebScoetClient(url)
	ws3, err := newWebScoetClient(url)
	if err != nil {
		t.Error(err)
		return
	}
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
	err = ws.WriteMessage(websocket.TextMessage, []byte("aaa"))
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
