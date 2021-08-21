package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/kcp"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().SetHandleKeyboardInterrupt(true)

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	go func() {
		errChan <- kcp.NewKcpListener(":8081", s, key, 4096).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start(dynamicStepHandler)
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
func dynamicStepHandler(step int, peers map[int]rua.Peer, msgs []rua.PeerMsg, s *rua.LockstepServer) (errs []error) {
	errs = []error{}

	if len(msgs) != 0 && len(msgs[0].Data) != 0 {
		clientTime := int64(binary.LittleEndian.Uint64(msgs[0].Data))
		serverTime := msgs[0].Time.UnixMilli()
		rtt := int(serverTime-clientTime) * 2 // round trip time

		fmt.Println("latency(ms):", serverTime-clientTime)
		s.SetStepLength(rtt)
		fmt.Println("new step length:", s.GetCurrentStepLength())
	}

	// write a blank package to go to the next step
	for _, p := range peers {
		if err := p.Write([]byte("0")); err != nil {
			errs = append(errs, err)
		}
	}
	return
}
