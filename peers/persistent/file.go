package persistent

import (
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

type filePeer struct {
	id       int // peer id
	fp       *os.File
	filename string
	gs       rua.GameServer
	lock     sync.Mutex
	tag      string
	logger   rua.Logger
}

func NewFilePeer(filename string, gs rua.GameServer) *filePeer {
	return &filePeer{
		gs:       gs,
		fp:       nil,
		filename: filename,
		lock:     sync.Mutex{},
		tag:      "file",
		logger:   rua.GetDefaultLogger(),
	}
}

func (p *filePeer) WithTag(t string) *filePeer {
	p.tag = t
	return p
}

func (p *filePeer) WithLogger(l rua.Logger) *filePeer {
	p.logger = l
	return p
}

func (p *filePeer) SetTag(t string) {
	p.tag = t
}

func (p *filePeer) GetTag() string {
	return p.tag
}

func (p *filePeer) Activate(id int) {
	p.id = id
}

// Thread safe.
func (p *filePeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := p.fp.Write(data)
	return err
}

func (p *filePeer) Close() error {
	// wait after write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.fp.Close() // close connection
}

func (p *filePeer) GetId() int {
	return p.id
}

func (p *filePeer) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()

	fp, err := os.OpenFile(p.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		p.logger.Error(err)
		return
	}
	p.fp = fp
}
