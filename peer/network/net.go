package network

import (
	"net"
	"time"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type NetPeer struct {
	*peer.BufferPeer
	closed       bool
	bufSize      int
	readTimeout  int
	writeTimeout int
	c            net.Conn
}

// Create a peer with a connection of `net.Conn`.
// If `timeout` == 0 (in ms), there is no timeout.
func NewNetPeer(conn net.Conn, gs rua.GameServer) *NetPeer {
	np := &NetPeer{
		BufferPeer:   peer.NewBufferPeer(gs),
		closed:       false,
		bufSize:      4096,
		readTimeout:  0,
		writeTimeout: 0,
		c:            conn,
	}

	np.BufferPeer.
		WithConsumer(func(data []byte) error {
			if np.closed {
				return rua.ErrPeerClosed
			}

			if np.writeTimeout != 0 {
				if err := np.c.SetWriteDeadline(time.Now().Add(time.Duration(np.readTimeout) * time.Millisecond)); err != nil {
					np.Logger().Error("rua.NetSetWriteDeadline:", err)
				}
			}
			_, err := np.c.Write(data)
			return err
		}).
		OnStartBuffer(func() {
			for {
				buf := make([]byte, np.bufSize)
				if np.readTimeout != 0 {
					if err := np.c.SetReadDeadline(time.Now().Add(time.Duration(np.readTimeout) * time.Millisecond)); err != nil {
						np.Logger().Error("rua.NetSetReadDeadline:", err)
					}
				}
				n, err := np.c.Read(buf)
				if err != nil {
					if !np.closed { // not closed by Close(), need to remove peer from server
						if err.Error() == "timeout" {
							np.Logger().Warnf("rua.NetPeer: peer[%d] timeout", np.Id())
						} else {
							np.Logger().Error("rua.NetOnStart:", err)
						}

						if err := gs.RemovePeer(np.Id()); err != nil {
							np.Logger().Error("rua.NetRemovePeer:", err)
						}
						break
					}
				} else {
					gs.AppendPeerMsg(np, buf[:n])
				}
			}
		}).
		OnCloseSafe(func() error {
			if np.closed {
				return nil
			}

			np.closed = true
			return np.c.Close()
		}).
		WithTag("net")

	return np
}

func (np *NetPeer) WithBufSize(n int) *NetPeer {
	np.bufSize = n
	return np
}

func (np *NetPeer) WithReadTimeout(ms int) *NetPeer {
	np.readTimeout = ms
	return np
}

func (np *NetPeer) WithWriteTimeout(ms int) *NetPeer {
	np.writeTimeout = ms
	return np
}

// Set the readTimeout and writeTimeout to the specified ms.
func (np *NetPeer) WithTimeout(ms int) *NetPeer {
	np.writeTimeout = ms
	np.readTimeout = ms
	return np
}
