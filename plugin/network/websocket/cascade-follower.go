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
}

func NewWebsocketCascadeFollower(leaderAddr string, gs rua.GameServer) *websocketCascadeFollower {
	return &websocketCascadeFollower{
		leaderAddr: leaderAddr,
		leaderPath: "/",
		gs:         gs,
	}
}

func (f *websocketCascadeFollower) WithLeaderPath(p string) *websocketCascadeFollower {
	f.leaderPath = p
	return f
}

func (f *websocketCascadeFollower) Connect() error {
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s%s", f.leaderAddr, f.leaderPath), nil)
	if err != nil {
		return err
	}
	f.gs.AddPeer(NewWebsocketPeer(c, f.gs))
	return nil
}
