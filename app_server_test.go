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

func (t testReceiver) Receive(tag string, s Sender, b []byte, data UserData) {

	s.SendAll(b)

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

		a := New("testKey")
		testlock.RLock()
		c, err := a.Register(tag[i], w, r, &testReceiver{}, nil)
		testlock.RUnlock()
		if err != nil {
			t.Error(err)
		}
		i++
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
	ap, err := Get("testKey")
	if err != nil {
		t.Error("empty key")
	}
	if ap.Count() != 3 {
		t.Error("count error", ap.Count())
	}
	if ap.CountById() != 2 {
		t.Error("count by tag error", ap.Count())
	}

}
