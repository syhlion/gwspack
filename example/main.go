package main

import (
	"github.com/syhlion/gwspack"
	"log"
	"net/http"
)

type Hello struct{}

func (h *Hello) Receive(s gwspack.Sender, b []byte) {
	s.SendAll(b)
}
func (h *Hello) GetUserData() gwspack.UserData {
	return nil
}
func main() {

	h := &Hello{}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("socket start")
		app := gwspack.Get("key")
		ws, err := app.Register("Frank", w, r, h)
		if err != nil {
			log.Println(err)
			log.Println("socket end")
			return
		}
		app.SendTo("Frank", []byte("hihi"))
		ws.Listen()
		log.Println("socket end")
		return
	})
	log.Fatal(http.ListenAndServe(":8888", nil))
}
