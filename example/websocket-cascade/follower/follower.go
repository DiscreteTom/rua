package main

import (
	"log"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
	ws "github.com/gorilla/websocket"
)

const leaderAddr = "ws://localhost:8080"

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(peers map[int]rua.Peer, m *rua.PeerMsg, s *rua.EventDrivenServer) {
			// print message from the leader
			log.Println(m.Data)
		})

	// connect to the leader
	c, _, err := ws.DefaultDialer.Dial(leaderAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
	s.AddPeer(websocket.NewWebsocketPeer(c, s))

	s.Start()
}
