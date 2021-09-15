package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
	"github.com/DiscreteTom/rua/peer/network"
	"github.com/DiscreteTom/rua/peer/persistent"
)

func main() {
	s := rua.NewEventDrivenServer()

	broadcaster := network.NewBroadcastPeer(s)

	s.OnPeerMsg(func(m *rua.PeerMsg) {
		broadcaster.Write(m.Data)
	})

	s.AddPeer(broadcaster)
	s.AddPeer(debug.NewStdioPeer(s))
	s.AddPeer(persistent.NewFilePeer("./output-1.txt", s))
	s.AddPeer(persistent.NewFilePeer("./output-2.txt", s))
	s.AddPeer(persistent.NewFilePeer("./output-3.txt", s))

	s.Start()
}
