package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/plugin/network/kcp"
	"DiscreteTom/rua/plugin/network/websocket"
	"crypto/sha1"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	s := lockstep.NewLockStepServer()

	ws := websocket.NewWebsocketListener(":8080", s)
	go ws.Start()

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	k := kcp.NewKcpListener(":8081", s, key, 4096)
	go k.Start()

	s.Start(broadcastStepHandler)
}
