package websocket

import (
	"github.com/DiscreteTom/rua"
)

func NewWebsocketCascadeLeader(addr string, gs rua.GameServer) *websocketListener {
	return NewWebsocketListener(addr, gs).WithPeerTag("websocket/cascade/leader")
}
