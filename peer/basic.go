package peer

import "github.com/DiscreteTom/rua"

type BasicPeer struct {
	id      int // peer id, assigned by game server
	gs      rua.GameServer
	tag     string
	logger  rua.Logger
	onWrite func(data []byte) error // lifecycle hook
	onClose func() error            // lifecycle hook
	onStart func()                  // lifecycle hook
}

type BasicPeerOption func(*BasicPeer) error

// Create a basic peer.
// Optional params: peer.Tag(), peer.Logger().
// You can use BasicPeer.OnWrite(), BasicPeer.OnClose(), BasicPeer.OnStart() to register lifecycle hooks.
func NewBasicPeer(gs rua.GameServer, options ...BasicPeerOption) (*BasicPeer, error) {
	p := &BasicPeer{
		gs:      gs,
		tag:     "basic",
		logger:  rua.DefaultLogger(),
		onWrite: func([]byte) error { return nil },
		onClose: func() error { return nil },
		onStart: func() {},
	}
	for _, option := range options {
		if err := option(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func Tag(t string) BasicPeerOption {
	return func(p *BasicPeer) error {
		p.tag = t
		return nil
	}
}

func Logger(l rua.Logger) BasicPeerOption {
	return func(p *BasicPeer) error {
		p.logger = l
		return nil
	}
}

// This should only be called when initializing a peer.
// Available options: peer.Tag(), peer.Logger().
func (p *BasicPeer) With(options ...BasicPeerOption) error {
	for _, option := range options {
		if err := option(p); err != nil {
			return err
		}
	}
	return nil
}

// This hook may be triggered concurrently
func (p *BasicPeer) OnWrite(f func(data []byte) error) {
	p.onWrite = f
}

// This hook may be triggered concurrently
func (p *BasicPeer) OnClose(f func() error) {
	p.onClose = f
}

// This hook may NOT be triggered concurrently
func (p *BasicPeer) OnStart(f func()) {
	p.onStart = f
}

func (p *BasicPeer) SetLogger(l rua.Logger) {
	p.logger = l
}

func (p *BasicPeer) Logger() rua.Logger {
	return p.logger
}

func (p *BasicPeer) SetTag(t string) {
	p.tag = t
}

func (p *BasicPeer) Tag() string {
	return p.tag
}

func (p *BasicPeer) SetId(id int) {
	p.id = id
}

func (p *BasicPeer) Id() int {
	return p.id
}

func (p *BasicPeer) GameServer() rua.GameServer {
	return p.gs
}

func (p *BasicPeer) Write(data []byte) error {
	return p.onWrite(data)
}

func (p *BasicPeer) Close() error {
	return p.onClose()
}

// Start and wait.
func (p *BasicPeer) Start() {
	p.onStart()
}
