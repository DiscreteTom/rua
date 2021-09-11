package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockstepServer()
	s.OnStep(broadcastStepHandler(s))

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

func broadcastStepHandler(s *rua.LockstepServer) func(msgs []rua.PeerMsg) {
	return func(msgs []rua.PeerMsg) {
		// compact msgs in one byte array
		result := []byte(fmt.Sprintf("step: %d\n", s.CurrentStep()))
		for _, msg := range msgs {
			result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.Peer.Id()))...)
			result = append(result, msg.Data...)
			result = append(result, '\n')
		}
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
