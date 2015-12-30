# Gusher WebScoket Pack
[![Build Status](https://travis-ci.org/syhlion/gwspack.svg?branch=master)](https://travis-ci.org/syhlion/gwspack)

Base on gorilla/websocket

## Install

`go get github.com/syhlion/gwspack`


## Useged

``` go
type Hello struct{}

func (h *Hello) Receive(s gwspack.Sender, b []byte) {
	s.SendAll(b)
}
func (h *Hello) GetUserData() UserData{
    return nil
}
func main() {

	h := &Hello{}

	app := gwspack.Get("key")
	go app.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		ws, err := app.Register("Frank", w, r,h)
		if err != nil {
			fmt.Println(err)
			return
		}
		ws.Listen()
		log.Println("socket end")
	})
	log.Fatal(http.ListenAndServe(":8888", nil))
}

```

