package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

const cascadeLeaderListenAddr = ":8080"

func main() {
	s := rua.NewEventDrivenServer()
	s.AfterAddPeer(func(newPeer rua.Peer) {
		// tell every cascade follower its leader peer id
		s.ForEachPeer(func(i int, p rua.Peer) {
			if p.Tag() == "websocket/cascade/leader" {
				p.Write([]byte(fmt.Sprintf("%d", i)))
			}
		})
	})

	go websocket.NewWebsocketCascadeLeader(cascadeLeaderListenAddr, s).Start()

	s.Start()
}
