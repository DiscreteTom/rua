package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer/debug"
)

func main() {
	s := rua.NewEventDrivenServer()
	s.OnPeerMsg(func(msg *rua.PeerMsg) {
		rua.WriteOrLog(msg.Peer, append(msg.Data, []byte("\n>>> ")...))
	})

	s.AddPeer(debug.NewStdioPeer(s))
	s.Start()
}
