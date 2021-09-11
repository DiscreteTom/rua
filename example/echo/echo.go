package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
)

func main() {
	s, _ := rua.NewEventDrivenServer()
	s.OnPeerMsg(func(msg *rua.PeerMsg) {
		if err := msg.Peer.Write(append([]byte(">>"), msg.Data...)); err != nil {
			s.Logger().Error(err)
		}
	})

	if p, err := debug.NewStdioPeer(s); err != nil {
		s.Logger().Error(err)
	} else {
		s.AddPeer(p)
	}
	s.Start()
}
