package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
	"github.com/DiscreteTom/rua/peer/persistent"
)

func main() {
	s := rua.NewEventDrivenServer()
	s.OnPeerMsg(func(msg *rua.PeerMsg) {
		s.ForEachPeer(func(id int, peer rua.Peer) {
			if peer.Tag() == "file" {
				peer.Write(msg.Data)
			}
		})
	})

	s.AddPeer(debug.NewStdioPeer(s))
	s.AddPeer(persistent.NewFilePeer("./log.txt", s))

	s.Start()
}
