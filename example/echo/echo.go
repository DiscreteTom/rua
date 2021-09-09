package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peers/debug"
)

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			if err := s.GetPeer(msg.PeerId).Write(msg.Data); err != nil {
				s.GetLogger().Error(err)
			}
		})

	if p, err := debug.NewStdioPeer(s); err != nil {
		s.AddPeer(p)
	} else {
		s.GetLogger().Error(err)
	}
	s.Start()
}
