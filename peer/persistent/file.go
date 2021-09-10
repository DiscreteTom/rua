package persistent

import (
	"os"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peer"
)

type FilePeer struct {
	*peer.SafePeer
	file     *os.File
	filename string // filename
}

func NewFilePeer(filename string, gs rua.GameServer, options ...peer.BasicPeerOption) (*FilePeer, error) {
	fp := &FilePeer{
		filename: filename,
	}

	sp, err := peer.NewSafePeer(
		gs,
		peer.Tag("file"),
	)
	if err != nil {
		return nil, err
	}

	sp.OnStartSafe(func() {
		var err error
		fp.file, err = os.OpenFile(fp.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fp.Logger().Error("rua.FileOpenFile:", err)
			return
		}
	})
	sp.OnWriteSafe(func(data []byte) error {
		if _, err := fp.file.Write(data); err != nil {
			return err
		}
		return fp.file.Sync() // flush to disk
	})
	sp.OnCloseSafe(func() error {
		return fp.file.Close() // close connection
	})

	fp.SafePeer = sp
	return fp, nil
}
