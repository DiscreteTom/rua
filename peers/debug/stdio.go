package debug

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

func NewStdioPeer(gs rua.GameServer) *rua.BasicPeer {
	lock := sync.Mutex{}

	return rua.NewBasicPeer(gs).
		WithTag("stdio").
		OnWrite(func(data []byte, p *rua.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			_, err := fmt.Print(string(data))
			return err
		}).
		OnClose(func(_ *rua.BasicPeer) error {
			// wait after write finished
			lock.Lock()
			defer lock.Unlock()

			return nil
		}).
		OnStart(func(p *rua.BasicPeer) {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					p.GetLogger().Error(err)
				}
				p.GetGameServer().AppendPeerMsg(p.GetId(), []byte(line))
			}
		})
}
