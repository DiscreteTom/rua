package main

import (
	"log"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

const leaderAddr = "localhost:8080"

func main() {
	s := rua.NewEventDrivenServer()
	s.OnPeerMsg(func(m *rua.PeerMsg) {
		if m.Peer.Tag() == "websocket/cascade/follower" {
			// print message from the leader
			log.Println(m.Data)
		}
	})

	// connect to the leader
	if err := websocket.NewWebsocketCascadeFollower(leaderAddr, s).Connect(); err != nil {
		log.Fatal(err)
	}

	s.Start()
}
