package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peers/debug"
	"github.com/DiscreteTom/rua/peers/persistent"
)

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			for _, p := range s.GetPeers() {
				if p.GetTag() == "file" {
					p.Write(msg.Data)
				}
			}
		})

	if p, err := debug.NewStdioPeer(s); err != nil {
		s.AddPeer(p)
	} else {
		s.GetLogger().Error(err)
	}
	if p, err := persistent.NewFilePeer("./log.txt", s); err != nil {
		s.AddPeer(p)
	} else {
		s.GetLogger().Error(err)
	}
	s.Start()
}
