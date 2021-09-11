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

// Create a basic peer.
// You can use BasicPeer.OnWrite(), BasicPeer.OnClose(), BasicPeer.OnStart() to register lifecycle hooks.
func NewBasicPeer(gs rua.GameServer) *BasicPeer {
	bp := &BasicPeer{
		gs:      gs,
		tag:     "basic",
		logger:  rua.DefaultLogger(),
		onWrite: func([]byte) error { return nil },
		onClose: func() error { return nil },
		onStart: func() {},
	}

	return bp
}

// This hook may be triggered concurrently
func (bp *BasicPeer) OnWrite(f func(data []byte) error) {
	bp.onWrite = f
}

// This hook may be triggered concurrently
func (bp *BasicPeer) OnClose(f func() error) {
	bp.onClose = f
}

// This hook may NOT be triggered concurrently
func (bp *BasicPeer) OnStart(f func()) {
	bp.onStart = f
}

func (bp *BasicPeer) SetLogger(l rua.Logger) {
	bp.logger = l
}

func (bp *BasicPeer) Logger() rua.Logger {
	return bp.logger
}

func (bp *BasicPeer) SetTag(t string) {
	bp.tag = t
}

func (bp *BasicPeer) Tag() string {
	return bp.tag
}

func (bp *BasicPeer) SetId(id int) {
	bp.id = id
}

func (bp *BasicPeer) Id() int {
	return bp.id
}

func (bp *BasicPeer) GameServer() rua.GameServer {
	return bp.gs
}

func (bp *BasicPeer) Write(data []byte) error {
	return bp.onWrite(data)
}

func (bp *BasicPeer) Close() error {
	return bp.onClose()
}

// Start and wait.
func (bp *BasicPeer) Start() {
	bp.onStart()
}
