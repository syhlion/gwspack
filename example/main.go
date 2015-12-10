package main

import (
	"fmt"
	"github.com/syhlion/wsexchange"
	"log"
	"net/http"
)

type Hello struct{}

func (h *Hello) Receive(tag string, s wsexchange.Sender, b []byte) {
	log.Println(tag)
	s.SendAll(b)
}
func main() {

	h := &Hello{}

	app := wsexchange.NewApp("key", h, 10)
	go app.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		ws, err := app.Register("Frank", w, r)
		if err != nil {
			fmt.Println(err)
			return
		}
		ws.Listen()
		log.Println("socket end")
	})
	log.Fatal(http.ListenAndServe(":8888", nil))
}
