package main

import (
	"encoding/binary"
	"time"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockstepServer()
	s.OnStep(dynamicStepHandler(s))

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

// Change step length according to the 1st msg's latency.
func dynamicStepHandler(s *rua.LockstepServer) func(msgs []rua.PeerMsg) {
	return func(msgs []rua.PeerMsg) {
		if len(msgs) != 0 && len(msgs[0].Data) == 8 {
			sendTime := int64(binary.LittleEndian.Uint64(msgs[0].Data))
			recvTime := msgs[0].Time.UnixMilli()
			rtt := int(recvTime - sendTime) // round trip time

			s.Logger().Info("rtt(ms):", rtt)
			s.WithStepLength(rtt)
			s.Logger().Info("new step length:", s.CurrentStepLength())
		}

		// broadcast current time
		buf := make([]byte, 8)
		currentTime := time.Now().UnixMilli()
		binary.LittleEndian.PutUint64(buf, uint64(currentTime))
		s.ForEachPeer(func(id int, peer rua.Peer) {
			if err := peer.Write(buf); err != nil {
				s.Logger().Error(err)
			}
		})
	}
}
