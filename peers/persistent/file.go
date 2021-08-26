package persistent

import (
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

func NewFilePeer(filename string, gs rua.GameServer) *rua.BasicPeer {
	lock := sync.Mutex{}
	var fp *os.File = nil

	return rua.NewBasicPeer(gs).
		WithTag("file").
		OnWrite(func(data []byte, p *rua.BasicPeer) error {
			// prevent concurrent write
			lock.Lock()
			defer lock.Unlock()

			if _, err := fp.Write(data); err != nil {
				return err
			}
			return fp.Sync() // flush to disk
		}).
		OnClose(func(p *rua.BasicPeer) error {
			// wait after write finished
			lock.Lock()
			defer lock.Unlock()

			return fp.Close() // close connection
		}).
		OnStart(func(p *rua.BasicPeer) {
			lock.Lock()
			defer lock.Unlock()

			var err error
			fp, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				p.GetLogger().Error(err)
				return
			}
		})
}
