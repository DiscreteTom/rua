package network

import (
	"errors"
	"net"
	"time"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type NetPeer struct {
	*peer.SafePeer
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
		closed:       false,
		bufSize:      4096,
		readTimeout:  0,
		writeTimeout: 0,
		c:            conn,
	}

	sp := peer.NewSafePeer(gs)
	sp.SetTag("net")

	sp.OnWriteSafe(func(data []byte) error {
		if !np.closed {
			if np.writeTimeout != 0 {
				if err := np.c.SetWriteDeadline(time.Now().Add(time.Duration(np.readTimeout) * time.Millisecond)); err != nil {
					np.Logger().Error("rua.NetSetWriteDeadline:", err)
				}
			}
			_, err := np.c.Write(data)
			return err
		}
		return errors.New("peer already closed")
	})
	sp.OnCloseSafe(func() error {
		np.closed = true
		return np.c.Close() // close connection
	})
	sp.OnStart(func() {
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
	})

	np.SafePeer = sp
	return np
}

func (np *NetPeer) SetBufSize(n int) {
	np.bufSize = n
}

func (np *NetPeer) SetReadTimeout(ms int) {
	np.readTimeout = ms
}

func (np *NetPeer) SetWriteTimeout(ms int) {
	np.writeTimeout = ms
}

// Set the readTimeout and writeTimeout to the specified ms.
func (np *NetPeer) SetTimeout(ms int) {
	np.writeTimeout = ms
}
