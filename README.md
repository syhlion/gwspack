# Gusher WebScoket Pack

Base on gorilla/websocket

## Install

`go get github.com/syhlion/gwspack`


## Useged

``` go
type Hello struct{}

func (h *Hello) Receive(tag string, s gwspack.Sender, b []byte, data gwspack.UserData) {
	log.Println(tag)
	s.SendAll(b)
}
func main() {

	h := &Hello{}

	app := gwspack.Get("key")
	go app.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		ws, err := app.Register("Frank", w, r,h,nil)
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

