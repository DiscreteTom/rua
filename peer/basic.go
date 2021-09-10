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

type BasicPeerOption func(*BasicPeer) error

// Create a basic peer.
// Optional params: peer.Tag(), peer.Logger(), peer.OnWrite(), peer.OnClose(), peer.OnStart()
func NewBasicPeer(gs rua.GameServer, options ...BasicPeerOption) (*BasicPeer, error) {
	p := &BasicPeer{
		gs:      gs,
		tag:     "basic",
		logger:  rua.DefaultLogger(),
		onWrite: func([]byte, *BasicPeer) error { return nil },
		onClose: func(*BasicPeer) error { return nil },
		onStart: func(*BasicPeer) {},
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
// Available options: peer.Tag(), peer.Logger(), peer.OnWrite(), peer.OnClose(), peer.OnStart()
func (p *BasicPeer) With(options ...BasicPeerOption) error {
	for _, option := range options {
		if err := option(p); err != nil {
			return err
		}
	}
	return nil
}

// This hook may be triggered concurrently
func OnWrite(f func(data []byte, p *BasicPeer) error) BasicPeerOption {
	return func(p *BasicPeer) error {
		p.onWrite = f
		return nil
	}
}

// This hook may be triggered concurrently
func OnClose(f func(p *BasicPeer) error) BasicPeerOption {
	return func(p *BasicPeer) error {
		p.onClose = f
		return nil
	}
}

// This hook may NOT be triggered concurrently
func OnStart(f func(p *BasicPeer)) BasicPeerOption {
	return func(p *BasicPeer) error {
		p.onStart = f
		return nil
	}
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
	return p.onWrite(data, p)
}

func (p *BasicPeer) Close() error {
	return p.onClose(p)
}

// Start and wait.
func (p *BasicPeer) Start() {
	p.onStart(p)
}
