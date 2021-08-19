package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/pkg/utils"
	"DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	s := lockstep.NewLockStepServer()

	ws := websocket.NewWebsocketListener(":8080", s)
	go ws.Start()

	s.Start(utils.BroadcastStepHandler)
}
