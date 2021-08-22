package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

const cascadeLeaderListenAddr = ":8080"

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		AfterAddPeer(func(newPeer rua.Peer, peers map[int]rua.Peer, s *rua.EventDrivenServer) {
			// tell every peer its id
			for i, p := range peers {
				p.Write([]byte(fmt.Sprintf("%d", i)))
			}
		})

	go websocket.NewWebsocketCascadeLeader(cascadeLeaderListenAddr, s).Start()

	s.Start()
}
