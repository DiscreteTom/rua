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
	p := &StdioPeer{}

	sp := peer.NewSafePeer(gs)

	sp.SetTag("stdio")

	sp.OnWriteSafe(func(data []byte) error {
		_, err := fmt.Print(string(data))
		return err
	})
	sp.OnCloseSafe(func() error {
		return nil
	})
	sp.OnStart(func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				p.Logger().Error("rua.StdioReadString:", err)
			}
			p.GameServer().AppendPeerMsg(p, []byte(line))
		}
	})

	p.SafePeer = sp
	return p
}
