package network

import (
	"net"
	"sync"
	"time"

	"github.com/DiscreteTom/rua"
)

// Create a peer with a connection of `net.Conn`.
// If `timeout` == 0 (in ms), there is no timeout.
func NewNetPeer(c net.Conn, gs rua.GameServer, bufSize int, readTimeout int, writeTimeout int) *rua.BasicPeer {
	lock := sync.Mutex{}
	closed := false

	return rua.NewBasicPeer(gs).
		WithTag("net").
		OnWrite(func(data []byte, p *rua.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			if writeTimeout != 0 {
				if err := c.SetWriteDeadline(time.Now().Add(time.Duration(readTimeout) * time.Millisecond)); err != nil {
					p.GetLogger().Error("rua.NetPeer.SetWriteDeadline:", err)
				}
			}
			_, err := c.Write(data)
			return err
		}).
		OnClose(func(p *rua.BasicPeer) error {
			// wait after write finished
			lock.Lock()
			defer lock.Unlock()

			closed = true
			return c.Close() // close connection
		}).
		OnStart(func(p *rua.BasicPeer) {
			for {
				buf := make([]byte, bufSize)
				if readTimeout != 0 {
					if err := c.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Millisecond)); err != nil {
						p.GetLogger().Error("rua.NetPeer.SetReadDeadline:", err)
					}
				}
				n, err := c.Read(buf)
				if err != nil {
					if closed {
						// closed by peer.Close(), not need to remove peer from server
						break
					}
					if err.Error() == "timeout" {
						p.GetLogger().Infof("rua.NetPeer: peer[%d] timeout", p.GetId())
					}
					if err := gs.RemovePeer(p.GetId()); err != nil {
						p.GetLogger().Error("rua.NetPeer.RemovePeer:", err)
					}
					break
				}

				gs.AppendPeerMsg(p.GetId(), buf[:n])
			}
		})
}
