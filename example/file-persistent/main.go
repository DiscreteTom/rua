package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peers/debug"
	"github.com/DiscreteTom/rua/peers/persistent"
)

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(peers map[int]rua.Peer, msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			for _, p := range peers {
				if p.GetTag() == "file" {
					p.Write(msg.Data)
				}
			}
		})

	s.AddPeer(debug.NewStdioPeer(s))
	s.AddPeer(persistent.NewFilePeer("./log.txt", s))
	s.Start()
}
