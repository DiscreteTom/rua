package websocket

import (
	"net/http"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type websocketListener struct {
	addr     string
	path     string
	gs       rua.GameServer
	guardian func(w http.ResponseWriter, r *http.Request, gs rua.GameServer) bool
	peerTag  string
	logger   rua.Logger
	certFile string
	keyFile  string
}

func NewWebsocketListener(addr string, gs rua.GameServer) *websocketListener {
	return &websocketListener{
		addr:     addr,
		path:     "/",
		gs:       gs,
		guardian: nil,
		peerTag:  "websocket",
		logger:   rua.GetDefaultLogger(),
		certFile: "",
		keyFile:  "",
	}
}

func (l *websocketListener) WithLogger(logger rua.Logger) *websocketListener {
	l.logger = logger
	return l
}

func (l *websocketListener) WithPath(p string) *websocketListener {
	l.path = p
	return l
}

func (l *websocketListener) WithPeerTag(t string) *websocketListener {
	l.peerTag = t
	return l
}

func (l *websocketListener) WithTLS(certFile, keyFile string) *websocketListener {
	l.certFile = certFile
	l.keyFile = keyFile
	return l
}

func (l *websocketListener) WithGuardian(g func(w http.ResponseWriter, r *http.Request, gs rua.GameServer) bool) *websocketListener {
	l.guardian = g
	return l
}

func (l *websocketListener) Start() error {
	http.HandleFunc(l.path, func(w http.ResponseWriter, r *http.Request) {
		if l.guardian == nil || l.guardian(w, r, l.gs) {
			// upgrade http to websocket
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				l.logger.Error(err)
				return
			}

			l.gs.AddPeer(NewWebsocketPeer(c, l.gs).WithLogger(l.logger).WithTag(l.peerTag))
		}
	})
	l.logger.Info("websocket listener is listening at", l.addr)

	if len(l.certFile) != 0 && len(l.keyFile) != 0 {
		return http.ListenAndServeTLS(l.addr, l.certFile, l.keyFile, nil)
	} else {
		return http.ListenAndServe(l.addr, nil)
	}
}
