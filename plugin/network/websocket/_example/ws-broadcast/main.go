package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

// Use `wscat -c ws://127.0.0.1:8080` to connect to the websocket server.
func main() {
	bc := rua.NewBroadcaster()

	ws, _ := websocket.NewWsListener("127.0.0.1:8080").OnNewPeer(func(wn *websocket.WsNode) {
		bc.AddTarget(wn.OnMsg(func(b []byte) {
			bc.Write(b)
		}).Go())
	}).Go()

	bc.AddTarget(rua.DefaultStdioNode().Go())

	rua.NewCtrlc().OnSignal(func() {
		bc.StopAll()
		ws.Stop()
	}).Wait()
}
