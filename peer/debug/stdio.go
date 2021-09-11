package debug

import (
	"bufio"
	"fmt"
	"os"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type StdioPeer struct {
	*peer.SafePeer
}

func NewStdioPeer(gs rua.GameServer) *StdioPeer {
	p := &StdioPeer{
		SafePeer: peer.NewSafePeer(gs),
	}

	p.SafePeer.
		OnWriteSafe(func(data []byte) error {
			_, err := fmt.Print(string(data))
			return err
		}).
		OnCloseSafe(func() error {
			return nil
		}).
		OnStart(func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					p.Logger().Error("rua.StdioReadString:", err)
				}
				p.GameServer().AppendPeerMsg(p, []byte(line))
			}
		}).
		WithTag("stdio")

	return p
}
