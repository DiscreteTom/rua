package main

import (
	"fmt"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(broadcastEventDrivenHandler)

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

func broadcastEventDrivenHandler(msg *rua.PeerMsg, s *rua.EventDrivenServer) {
	// compact msg in one byte array
	result := []byte{}
	result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.PeerId))...)
	result = append(result, msg.Data...)
	result = append(result, '\n')
	// broadcast to everyone
	for _, p := range s.GetPeers() {
		go func() {
			if err := p.Write(result); err != 0 {
				s.GetLogger().Error(err)
			}
		}()
	}
}
