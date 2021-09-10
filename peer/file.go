package peer

import (
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

type FilePeer struct {
	BasicPeer
	lock *sync.Mutex
	fp   *os.File
	fn   string // filename
}

func NewFilePeer(filename string, gs rua.GameServer) (*FilePeer, error) {
	p := &FilePeer{
		lock: &sync.Mutex{},
		fn:   filename,
	}

	bp, err := NewBasicPeer(
		gs,
		Tag("file"),
		OnWrite(func(data []byte, _ *BasicPeer) error {
			// prevent concurrent write
			p.lock.Lock()
			defer p.lock.Unlock()

			if _, err := p.fp.Write(data); err != nil {
				return err
			}
			return p.fp.Sync() // flush to disk
		}),
		OnClose(func(_ *BasicPeer) error {
			// wait after write finished
			p.lock.Lock()
			defer p.lock.Unlock()

			return p.fp.Close() // close connection
		}),
		OnStart(func(_ *BasicPeer) {
			p.lock.Lock()
			defer p.lock.Unlock()

			var err error
			p.fp, err = os.OpenFile(p.fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				p.Logger().Error("rua.FileOpenFile:", err)
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
