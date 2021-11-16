package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

// Use `wscat -c ws://127.0.0.1:8080` to connect to the websocket server.
func main() {
	// create broadcaster
	bc := rua.NewBroadcaster()

	// websocket listener
	ws, _ := websocket.NewWsListener("127.0.0.1:8080").OnNewPeer(func(wn *websocket.WsNode) {
		// add new peer to the broadcaster
		bc.AddTarget(
			wn.OnMsg(
				func(b []byte) {
					// new message will be write to the broadcaster
					bc.Write(b)
				},
			).Go(),
		)
	}).Go() // start the listener

	// wait for ctrl-c
	rua.NewCtrlc().OnSignal(func() {
		bc.StopAll()
		ws.Stop()
	}).Wait()
}
