package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	s := lockstep.NewLockStepServer(30, lockstep.Broadcast)

	ws := websocket.NewWebsocketListener(":8080", s)
	go ws.Start()

	s.Start()
}
