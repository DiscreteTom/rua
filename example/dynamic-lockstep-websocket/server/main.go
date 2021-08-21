package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().
		SetHandleKeyboardInterrupt(true).
		On(rua.Step, dynamicStepHandler)

	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		log.Println(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			log.Println(errs)
		}
		break
	}
}

// Change step length according to the 1st msg's latency.
func dynamicStepHandler(step int, peers map[int]rua.Peer, msgs []rua.PeerMsg, s *rua.LockstepServer) {
	if len(msgs) != 0 && len(msgs[0].Data) == 8 {
		sendTime := int64(binary.LittleEndian.Uint64(msgs[0].Data))
		recvTime := msgs[0].Time.UnixMilli()
		rtt := int(recvTime - sendTime) // round trip time

		fmt.Println("rtt(ms):", rtt)
		s.SetStepLength(rtt)
		fmt.Println("new step length:", s.GetCurrentStepLength())
	}

	// broadcast current time
	buf := make([]byte, 8)
	currentTime := time.Now().UnixMilli()
	binary.LittleEndian.PutUint64(buf, uint64(currentTime))
	for _, p := range peers {
		if err := p.Write(buf); err != nil {
			log.Println(err)
		}
	}
}
