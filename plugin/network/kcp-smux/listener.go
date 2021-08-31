package kcpsmux

import (
	"net"

	"github.com/DiscreteTom/rua"
	ruaKcp "github.com/DiscreteTom/rua/plugin/network/kcp"
	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

// If upgrader fails, it will return the original kcp.UDPSession and an error.
func Upgrader(c *kcp.UDPSession) (net.Conn, error) {
	// setup smux server
	session, err := smux.Server(c, nil)
	if err != nil {
		return c, err
	} else {
		// accept as a stream
		stream, err := session.AcceptStream()
		if err != nil {
			return c, err
		} else {
			return stream, nil
		}
	}
}

func NewKcpSmuxListener(addr string, gs rua.GameServer, key []byte, bufSize int) *ruaKcp.KcpListener {
	return ruaKcp.NewKcpListener(addr, gs, key, bufSize).WithPeerTag("kcpsmux").WithUpgrader(Upgrader)
}
