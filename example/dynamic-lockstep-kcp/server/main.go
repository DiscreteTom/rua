package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/pkg/model"
	"DiscreteTom/rua/plugin/network/kcp"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	errChan := make(chan error)
	s := lockstep.NewLockStepServer().SetHandleKeyboardInterrupt(true)

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	k := kcp.NewKcpListener(":8081", s, key, 4096)
	go func() {
		errChan <- k.Start()
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

// Change step length according to the 1st peer's latency.
func dynamicStepHandler(step int, peers map[int]model.Peer, commands map[int][]model.PeerCommand, s *lockstep.LockstepServer) (errs []error) {
	errs = []error{}
	if p, ok := peers[0]; ok {
		if len(commands[0]) != 0 && len(commands[0][0].Data) != 0 {
			clientTime := int64(binary.LittleEndian.Uint64(commands[0][0].Data))
			serverTime := commands[0][0].Time.UnixMilli()
			rtt := int(serverTime-clientTime) * 2 // round trip time

			fmt.Println("latency(ms):", serverTime-clientTime)
			s.SetStepLength(rtt)
			fmt.Println("new step length:", s.GetCurrentStepLength())
		}

		// write a blank package to go to the next step
		if err := p.Write([]byte("00000000")); err != nil {
			errs = append(errs, err)
		}
	}
	return
}
