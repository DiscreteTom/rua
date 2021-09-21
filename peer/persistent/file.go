package persistent

import (
	"errors"
	"os"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type FilePeer struct {
	*peer.SafePeer
	file     *os.File
	closed   bool
	filename string // filename
}

func NewFilePeer(filename string, gs rua.GameServer) *FilePeer {
	fp := &FilePeer{
		closed:   true,
		SafePeer: peer.NewSafePeer(gs),
		filename: filename,
	}

	fp.SafePeer.
		OnStartSafe(func() {
			var err error
			fp.file, err = os.OpenFile(fp.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				fp.Logger().Error("rua.FileOpenFile:", err)
			} else {
				fp.closed = false
			}
		}).
		OnWriteSafe(func(data []byte) error {
			if fp.closed {
				return rua.ErrPeerClosed
			}

			if _, err := fp.file.Write(data); err != nil {
				return err
			}
			return fp.file.Sync() // flush to disk
		}).
		OnCloseSafe(func() error {
			if fp.closed {
				return nil
			}

			if err := fp.file.Close(); errors.Is(err, os.ErrClosed) {
				return nil
			} else {
				return err
			}
		}).
		WithTag("file")

	return fp
}
