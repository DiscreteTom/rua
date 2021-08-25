package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peers/debug"
)

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(peers map[int]rua.Peer, msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			if err := peers[msg.PeerId].Write(msg.Data); err != nil {
				s.GetLogger().Error(err)
			}
		})

	s.AddPeer(debug.NewStdioPeer(s))
	s.Start()
}
