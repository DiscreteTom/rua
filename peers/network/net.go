package network

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/DiscreteTom/rua"
	peer "github.com/DiscreteTom/rua/peers/basic"
)

type NetPeer struct {
	peer.BasicPeer
	lock   *sync.Mutex
	closed bool
}

// Create a peer with a connection of `net.Conn`.
// If `timeout` == 0 (in ms), there is no timeout.
func NewNetPeer(c net.Conn, gs rua.GameServer, bufSize int, readTimeout int, writeTimeout int) (*NetPeer, error) {
	p := &NetPeer{
		lock:   &sync.Mutex{},
		closed: false,
	}

	bp, err := peer.NewBasicPeer(
		gs,
		peer.Tag("net"),
		peer.OnWrite(func(data []byte, bp *peer.BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			if !p.closed {
				if writeTimeout != 0 {
					if err := c.SetWriteDeadline(time.Now().Add(time.Duration(readTimeout) * time.Millisecond)); err != nil {
						bp.GetLogger().Error("rua.NetPeer.SetWriteDeadline:", err)
					}
				}
				_, err := c.Write(data)
				return err
			}
			return errors.New("peer already closed")
		}),
		peer.OnClose(func(bp *peer.BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			p.closed = true
			return c.Close() // close connection
		}),
		peer.OnStart(func(bp *peer.BasicPeer) {
			for {
				buf := make([]byte, bufSize)
				if readTimeout != 0 {
					if err := c.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Millisecond)); err != nil {
						p.GetLogger().Error("rua.NetPeer.SetReadDeadline:", err)
					}
				}
				n, err := c.Read(buf)
				if err != nil {
					if !p.closed { // not closed by peer.Close(), need to remove peer from server
						if err.Error() == "timeout" {
							p.GetLogger().Warnf("rua.NetPeer: peer[%d] timeout", p.GetId())
						} else {
							p.GetLogger().Error("rua.NetPeer.OnStart:", err)
						}

						if err := gs.RemovePeer(p.GetId()); err != nil {
							p.GetLogger().Error("rua.NetPeer.RemovePeer:", err)
						}
						break
					}
				} else {
					gs.AppendPeerMsg(p.GetId(), buf[:n])
				}
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	p.BasicPeer = *bp
	return p, nil
}
