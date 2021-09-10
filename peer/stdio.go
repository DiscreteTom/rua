package peer

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

type StdioPeer struct {
	BasicPeer
	lock *sync.Mutex
}

func NewStdioPeer(gs rua.GameServer) (*StdioPeer, error) {
	p := &StdioPeer{lock: &sync.Mutex{}}

	bp, err := NewBasicPeer(
		gs,
		Tag("stdio"),
		OnWrite(func(data []byte, _ *BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			_, err := fmt.Print(string(data))
			return err
		}),
		OnClose(func(_ *BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			return nil
		}),
		OnStart(func(_ *BasicPeer) {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					p.Logger().Error("rua.StdioReadString:", err)
				}
				p.GameServer().AppendPeerMsg(p.Id(), []byte(line))
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	p.BasicPeer = *bp
	return p, nil
}
