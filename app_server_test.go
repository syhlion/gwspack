package wsexchange

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testReceiver struct{ *testing.T }

func (t testReceiver) Receive(tag string, s Sender, b []byte) {

	s.SendAll(b)

}

func newApp(t *testing.T) *App {

	var app *App
	app = NewApp("testKey", &testReceiver{t}, 5)
	if app.key != "testKey" {
		t.Error("error key")
	}
	return app
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
	app := newApp(t)
	go app.Run()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("test")
		c, err := app.Register("Frank", w, r)
		if err != nil {
			t.Error(err)
		}
		c.Listen()

	}))
	defer ts.Close()
	ws := newWebScoetClient(ts.URL)
	ws2 := newWebScoetClient(ts.URL)
	defer func() {
		ws.Close()
		ws2.Close()
	}()
	if app.Count() != 2 {
		t.Error("not in map")
	}
	if len(app.receiverProcessPool) != 1 {
		t.Error("process not stratr")
	}

}
