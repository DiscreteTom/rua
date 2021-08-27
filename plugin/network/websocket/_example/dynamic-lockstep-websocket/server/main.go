package main

import (
	"encoding/binary"
	"time"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().
		SetHandleKeyboardInterrupt(true).
		OnStep(dynamicStepHandler)

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

// Change step length according to the 1st msg's latency.
func dynamicStepHandler(msgs []rua.PeerMsg, s *rua.LockstepServer) {
	if len(msgs) != 0 && len(msgs[0].Data) == 8 {
		sendTime := int64(binary.LittleEndian.Uint64(msgs[0].Data))
		recvTime := msgs[0].Time.UnixMilli()
		rtt := int(recvTime - sendTime) // round trip time

		s.GetLogger().Info("rtt(ms):", rtt)
		s.SetStepLength(rtt)
		s.GetLogger().Info("new step length:", s.GetCurrentStepLength())
	}

	// broadcast current time
	buf := make([]byte, 8)
	currentTime := time.Now().UnixMilli()
	binary.LittleEndian.PutUint64(buf, uint64(currentTime))
	for _, p := range s.GetPeers() {
		if err := p.Write(buf); err != nil {
			s.GetLogger().Error(err)
		}
	}
}
