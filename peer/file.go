package peer

import (
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
	peer "github.com/DiscreteTom/rua/peers/basic"
)

type FilePeer struct {
	peer.BasicPeer
	lock *sync.Mutex
	fp   *os.File
	fn   string // filename
}

func NewFilePeer(filename string, gs rua.GameServer) (*FilePeer, error) {
	p := &FilePeer{
		lock: &sync.Mutex{},
		fn:   filename,
	}

	bp, err := peer.NewBasicPeer(
		gs,
		peer.Tag("file"),
		peer.OnWrite(func(data []byte, _ *peer.BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			if _, err := p.fp.Write(data); err != nil {
				return err
			}
			return p.fp.Sync() // flush to disk
		}),
		peer.OnClose(func(_ *peer.BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			return p.fp.Close() // close connection
		}),
		peer.OnStart(func(_ *peer.BasicPeer) {
			p.lock.Lock()
			defer p.lock.Unlock()

			var err error
			p.fp, err = os.OpenFile(p.fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				p.GetLogger().Error("rua.FilePeer.OpenFile:", err)
				return
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	p.BasicPeer = *bp
	return p, nil
}
