package main

import (
	"DiscreteTom/rua/pkg/lockstep"
	"DiscreteTom/rua/pkg/model"
	"DiscreteTom/rua/plugin/network/kcp"
	"crypto/sha1"
	"encoding/binary"
	"log"
	"time"

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
func dynamicStepHandler(step int, peers map[int]model.Peer, commands map[int][]byte, s *lockstep.LockstepServer) (errs []error) {
	errs = []error{}
	if p, ok := peers[0]; ok {
		if len(commands[0]) != 0 {
			clientTime := int64(binary.LittleEndian.Uint64(commands[0]))
			serverTime := time.Now().UnixMilli()
			newStepLength := int(serverTime-clientTime) * 2

			log.Println("step:", step, "now:", serverTime, "latency(ms):", serverTime-clientTime, "newStepLength:", newStepLength)
			s.SetStepLength(newStepLength)
		}

		// write a blank package to go to the next step
		if err := p.Write([]byte("0")); err != nil {
			errs = append(errs, err)
		}
	}
	return
}
