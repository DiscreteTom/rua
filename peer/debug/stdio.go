package debug

import (
	"bufio"
	"fmt"
	"os"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type StdioPeer struct {
	*peer.BufferPeer
}

func NewStdioPeer(gs rua.GameServer) *StdioPeer {
	p := &StdioPeer{
		BufferPeer: peer.NewBufferPeer(gs),
	}

	p.BufferPeer.
		WithConsumer(func(data []byte) error {
			_, err := fmt.Print(string(data))
			return err
		}).
		OnStartBuffer(func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					p.Logger().Error("rua.StdioReadString:", err)
				}
				p.GameServer().AppendPeerMsg(p, []byte(line))
			}
		}).
		OnCloseSafe(func() error {
			return nil
		}).
		WithTag("stdio")

	return p
}
