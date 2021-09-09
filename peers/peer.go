package peer

import "github.com/DiscreteTom/rua"

type BasicPeer struct {
	id      int // peer id, assigned by game server
	gs      rua.GameServer
	tag     string
	logger  rua.Logger
	onWrite func(data []byte, p *BasicPeer) error // lifecycle hook
	onClose func(p *BasicPeer) error              // lifecycle hook
	onStart func(p *BasicPeer)                    // lifecycle hook
}

// Create a basic peer.
// You should call `WithTag`/`OnWrite`/`OnClose`/`OnStart` after this call.
func NewBasicPeer(gs rua.GameServer) *BasicPeer {
	return &BasicPeer{
		gs:      gs,
		tag:     "basic",
		logger:  rua.GetDefaultLogger(),
		onWrite: func([]byte, *BasicPeer) error { return nil },
		onClose: func(*BasicPeer) error { return nil },
		onStart: func(*BasicPeer) {},
	}
}

func (p *BasicPeer) WithTag(t string) *BasicPeer {
	p.tag = t
	return p
}

func (p *BasicPeer) WithLogger(l rua.Logger) *BasicPeer {
	p.logger = l
	return p
}

// This hook may be triggered concurrently
func (p *BasicPeer) OnWrite(f func(data []byte, p *BasicPeer) error) *BasicPeer {
	p.onWrite = f
	return p
}

// This hook may be triggered concurrently
func (p *BasicPeer) OnClose(f func(p *BasicPeer) error) *BasicPeer {
	p.onClose = f
	return p
}

// This hook may be triggered concurrently
func (p *BasicPeer) OnStart(f func(p *BasicPeer)) *BasicPeer {
	p.onStart = f
	return p
}

func (p *BasicPeer) SetLogger(l rua.Logger) {
	p.logger = l
}

func (p *BasicPeer) GetLogger() rua.Logger {
	return p.logger
}

func (p *BasicPeer) SetTag(t string) {
	p.tag = t
}

func (p *BasicPeer) GetTag() string {
	return p.tag
}

func (p *BasicPeer) SetId(id int) {
	p.id = id
}

func (p *BasicPeer) GetId() int {
	return p.id
}

func (p *BasicPeer) GetGameServer() rua.GameServer {
	return p.gs
}

// Thread safe.
func (p *BasicPeer) Write(data []byte) error {
	return p.onWrite(data, p)
}

func (p *BasicPeer) Close() error {
	return p.onClose(p)
}

// Start and wait.
func (p *BasicPeer) Start() {
	p.onStart(p)
}
