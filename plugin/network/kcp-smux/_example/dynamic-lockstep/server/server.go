package main

import (
	"crypto/sha1"
	"encoding/binary"
	"time"

	"github.com/DiscreteTom/rua"
	kcpsmux "github.com/DiscreteTom/rua/plugin/network/kcp-smux"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	errChan := make(chan error)
	s := rua.NewLockStepServer().
		SetHandleKeyboardInterrupt(true).
		OnStep(dynamicStepHandler)

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	go func() {
		errChan <- kcpsmux.NewKcpSmuxListener(":8081", s, key, 4096).Start()
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

		s.GetLogger().Infof("rtt(ms): %d", rtt)
		s.SetStepLength(rtt)
		s.GetLogger().Infof("new step length: %d", s.GetCurrentStepLength())
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
