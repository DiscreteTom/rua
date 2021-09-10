package peer

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
	lock         *sync.Mutex
	closed       bool
	bufSize      int
	readTimeout  int
	writeTimeout int
	c            net.Conn
}

type NetPeerOption func(*NetPeer) error

// Create a peer with a connection of `net.Conn`.
// If `timeout` == 0 (in ms), there is no timeout.
func NewNetPeer(connection net.Conn, gs rua.GameServer, options ...NetPeerOption) (*NetPeer, error) {
	p := &NetPeer{
		lock:         &sync.Mutex{},
		closed:       false,
		bufSize:      4096,
		readTimeout:  0,
		writeTimeout: 0,
		c:            connection,
	}

	bp, err := peer.NewBasicPeer(
		gs,
		peer.Tag("net"),
		peer.OnWrite(func(data []byte, _ *peer.BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			if !p.closed {
				if p.writeTimeout != 0 {
					if err := p.c.SetWriteDeadline(time.Now().Add(time.Duration(p.readTimeout) * time.Millisecond)); err != nil {
						p.GetLogger().Error("rua.NetPeer.SetWriteDeadline:", err)
					}
				}
				_, err := p.c.Write(data)
				return err
			}
			return errors.New("peer already closed")
		}),
		peer.OnClose(func(_ *peer.BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			p.closed = true
			return p.c.Close() // close connection
		}),
		peer.OnStart(func(_ *peer.BasicPeer) {
			for {
				buf := make([]byte, p.bufSize)
				if p.readTimeout != 0 {
					if err := p.c.SetReadDeadline(time.Now().Add(time.Duration(p.readTimeout) * time.Millisecond)); err != nil {
						p.GetLogger().Error("rua.NetPeer.SetReadDeadline:", err)
					}
				}
				n, err := p.c.Read(buf)
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

func BufSize(n int) NetPeerOption {
	return func(p *NetPeer) error {
		p.bufSize = n
		return nil
	}
}

func ReadTimeout(ms int) NetPeerOption {
	return func(p *NetPeer) error {
		p.readTimeout = ms
		return nil
	}
}

func WriteTimeout(ms int) NetPeerOption {
	return func(p *NetPeer) error {
		p.writeTimeout = ms
		return nil
	}
}

// Set the readTimeout and writeTimeout to the specified ms.
func Timeout(ms int) NetPeerOption {
	return func(p *NetPeer) error {
		p.writeTimeout = ms
		p.readTimeout = ms
		return nil
	}
}
