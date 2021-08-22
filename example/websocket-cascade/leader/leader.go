package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
	ws "github.com/gorilla/websocket"
)

const wsAddr = ":8080"

var upgrader = ws.Upgrader{}

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		AfterAddPeer(func(newPeer rua.Peer, peers map[int]rua.Peer, s *rua.EventDrivenServer) {
			// tell every peer its id
			for i, p := range peers {
				p.Write([]byte(fmt.Sprintf("%d", i)))
			}
		})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		s.AddPeer(websocket.NewWebsocketPeer(c, s))
	})
	log.Println("websocket server is listening at", wsAddr)
	go http.ListenAndServe(wsAddr, nil)

	s.Start()
}
