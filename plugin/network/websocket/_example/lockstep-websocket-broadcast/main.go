package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().
		SetHandleKeyboardInterrupt(true).
		OnStep(broadcastStepHandler)

	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		s.GetLogger().Error(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			s.GetLogger().Error(errs)
		}
		break
	}
}

func broadcastStepHandler(step int, peers map[int]rua.Peer, msgs []rua.PeerMsg, s *rua.LockstepServer) {
	// compact msgs in one byte array
	result := []byte(fmt.Sprintf("step: %d\n", step))
	for _, msg := range msgs {
		result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.PeerId))...)
		result = append(result, msg.Data...)
		result = append(result, '\n')
	}
	// broadcast to everyone
	for _, p := range peers {
		go func() {
			if err := p.Write(result); err != nil {
				s.GetLogger().Error(err)
			}
		}()
	}
}
