package websocket

import (
	"errors"
	"net/http"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type wsListener struct {
	addr            string
	path            string
	guardian        func(w http.ResponseWriter, r *http.Request) bool
	certFile        string
	keyFile         string
	upgrader        *websocket.Upgrader
	peerWriteBuffer uint
	handle          *rua.StopOnlyHandle
	stopRx          chan *rua.StopPayload
	peerHandler     func(*WsNode)
}

func NewWsListener(addr string) *wsListener {
	stopChan := make(chan *rua.StopPayload)
	handle, _ := rua.NewHandleBuilder().StopTx(stopChan).BuildStopOnly()

	return &wsListener{
		addr:            addr,
		path:            "/",
		guardian:        nil,
		certFile:        "",
		keyFile:         "",
		upgrader:        &websocket.Upgrader{},
		peerWriteBuffer: 16,
		handle:          handle,
		stopRx:          stopChan,
		peerHandler:     nil,
	}
}

func (l *wsListener) OnNewPeer(f func(*WsNode)) *wsListener {
	l.peerHandler = f
	return l
}

func (l *wsListener) PeerWriteBuffer(buffer uint) *wsListener {
	l.peerWriteBuffer = buffer
	return l
}

func (l *wsListener) Path(p string) *wsListener {
	l.path = p
	return l
}

func (l *wsListener) TLS(certFile, keyFile string) *wsListener {
	l.certFile = certFile
	l.keyFile = keyFile
	return l
}

func (l *wsListener) Guardian(g func(w http.ResponseWriter, r *http.Request) bool) *wsListener {
	l.guardian = g
	return l
}

func (l *wsListener) OriginChecker(f func(r *http.Request) bool) *wsListener {
	l.upgrader.CheckOrigin = f
	return l
}

func (l *wsListener) AllowAnyOrigin() *wsListener {
	l.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return l
}

func (l *wsListener) Handle() *rua.StopOnlyHandle {
	return l.handle
}

func (l *wsListener) Go() (*rua.StopOnlyHandle, error) {
	if l.peerHandler == nil {
		return nil, errors.New("missing peerHandler")
	}

	go func() {
		http.HandleFunc(l.path, func(w http.ResponseWriter, r *http.Request) {
			if l.guardian == nil || l.guardian(w, r) {
				// upgrade http to websocket
				c, err := l.upgrader.Upgrade(w, r, nil)
				if err != nil {
					panic(err)
				}

				l.peerHandler(NewWsNode(c, l.peerWriteBuffer))
			}
		})
		if len(l.certFile) != 0 && len(l.keyFile) != 0 {
			http.ListenAndServeTLS(l.addr, l.certFile, l.keyFile, nil)
		} else {
			http.ListenAndServe(l.addr, nil)
		}
	}()

	return l.handle, nil
}
