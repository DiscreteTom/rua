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
	s := rua.NewLockstepServer()
	s.OnStep(dynamicStepHandler(s))

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
		s.Logger().Error("server.KcpSmuxListener:", err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			s.Logger().Error("server.LockstepServer:", errs)
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

			s.Logger().Infof("rtt(ms): %d", rtt)
			s.WithStepLength(rtt)
			s.Logger().Infof("new step length: %d", s.CurrentStepLength())
		}

		// broadcast current time
		buf := make([]byte, 8)
		currentTime := time.Now().UnixMilli()
		binary.LittleEndian.PutUint64(buf, uint64(currentTime))
		s.ForEachPeer(func(_ int, p rua.Peer) {
			if err := p.Write(buf); err != nil {
				s.Logger().Error("server.Broadcast.Write:", err)
			}
		})
	}
}
