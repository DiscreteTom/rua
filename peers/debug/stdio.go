package debug

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
	peer "github.com/DiscreteTom/rua/peers/basic"
)

func NewStdioPeer(gs rua.GameServer) (*peer.BasicPeer, error) {
	lock := sync.Mutex{}

	return peer.NewBasicPeer(
		gs,
		peer.Tag("stdio"),
		peer.OnWrite(func(data []byte, p *peer.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			_, err := fmt.Print(string(data))
			return err
		}),
		peer.OnClose(func(_ *peer.BasicPeer) error {
			// wait after write finished
			lock.Lock()
			defer lock.Unlock()

			return nil
		}),
		peer.OnStart(func(p *peer.BasicPeer) {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					p.GetLogger().Error("rua.StdioPeer.ReadString:", err)
				}
				p.GetGameServer().AppendPeerMsg(p.GetId(), []byte(line))
			}
		}),
	)
}
