package peer

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
	peer "github.com/DiscreteTom/rua/peers/basic"
)

type StdioPeer struct {
	peer.BasicPeer
	lock *sync.Mutex
}

func NewStdioPeer(gs rua.GameServer) (*StdioPeer, error) {
	p := &StdioPeer{lock: &sync.Mutex{}}

	bp, err := peer.NewBasicPeer(
		gs,
		peer.Tag("stdio"),
		peer.OnWrite(func(data []byte, _ *peer.BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			_, err := fmt.Print(string(data))
			return err
		}),
		peer.OnClose(func(_ *peer.BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			return nil
		}),
		peer.OnStart(func(_ *peer.BasicPeer) {
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
	if err != nil {
		return nil, err
	}

	p.BasicPeer = *bp
	return p, nil
}
