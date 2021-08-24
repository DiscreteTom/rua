package websocket

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/gorilla/websocket"
)

type websocketCascadeFollower struct {
	leaderAddr string
	leaderPath string
	gs         rua.GameServer
	peerTag    string
}

func NewWebsocketCascadeFollower(leaderAddr string, gs rua.GameServer) *websocketCascadeFollower {
	return &websocketCascadeFollower{
		leaderAddr: leaderAddr,
		leaderPath: "/",
		gs:         gs,
		peerTag:    "websocket/cascade/follower",
	}
}

func (f *websocketCascadeFollower) WithPeerTag(t string) *websocketCascadeFollower {
	f.peerTag = t
	return f
}

func (f *websocketCascadeFollower) WithLeaderPath(p string) *websocketCascadeFollower {
	f.leaderPath = p
	return f
}

// Connect to the cascade leader & add a peer to the game server.
// Return websocket dial error.
func (f *websocketCascadeFollower) Connect() error {
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s%s", f.leaderAddr, f.leaderPath), nil)
	if err != nil {
		return err
	}
	f.gs.AddPeer(NewWebsocketPeer(c, f.gs).WithTag(f.peerTag))
	return nil
}
