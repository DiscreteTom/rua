package debug

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/DiscreteTom/rua"
)

type stdioPeer struct {
	id     int // peer id
	gs     rua.GameServer
	lock   sync.Mutex
	tag    string
	logger rua.Logger
}

func NewStdioPeer(gs rua.GameServer) *stdioPeer {
	return &stdioPeer{
		gs:     gs,
		lock:   sync.Mutex{},
		tag:    "stdio",
		logger: rua.GetDefaultLogger(),
	}
}

func (p *stdioPeer) WithTag(t string) *stdioPeer {
	p.tag = t
	return p
}

func (p *stdioPeer) WithLogger(l rua.Logger) *stdioPeer {
	p.logger = l
	return p
}

func (p *stdioPeer) SetTag(t string) {
	p.tag = t
}

func (p *stdioPeer) GetTag() string {
	return p.tag
}

func (p *stdioPeer) Activate(id int) {
	p.id = id
}

// Thread safe.
func (p *stdioPeer) Write(data []byte) error {
	// prevent concurrent write
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err := fmt.Print(string(data))
	return err
}

func (p *stdioPeer) Close() error {
	// wait after write finished
	p.lock.Lock()
	defer p.lock.Unlock()

	return nil
}

func (p *stdioPeer) GetId() int {
	return p.id
}

func (p *stdioPeer) Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			p.logger.Error(err)
		}
		p.gs.AppendPeerMsg(p.id, []byte(line))
	}
}
