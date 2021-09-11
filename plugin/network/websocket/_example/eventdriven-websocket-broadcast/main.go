package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewEventDrivenServer()
	s.OnPeerMsg(broadcastEventDrivenHandler(s))

	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		s.Logger().Error(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			s.Logger().Error(errs)
		}
		break
	}
}

func broadcastEventDrivenHandler(s *rua.EventDrivenServer) func(msg *rua.PeerMsg) {
	return func(msg *rua.PeerMsg) {
		// compact msg in one byte array
		result := []byte{}
		result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.Peer.Id()))...)
		result = append(result, msg.Data...)
		result = append(result, '\n')
		// broadcast to everyone
		s.ForEachPeer(func(id int, peer rua.Peer) {
			go func() {
				if err := peer.Write(result); err != nil {
					s.Logger().Error(err)
				}
			}()
		})
	}
}
