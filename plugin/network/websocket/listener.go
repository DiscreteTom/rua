package websocket

import (
	"net/http"

	"github.com/DiscreteTom/rua"

	"github.com/gorilla/websocket"
)

type websocketListener struct {
	name     string
	addr     string
	path     string
	gs       rua.GameServer
	guardian func(w http.ResponseWriter, r *http.Request) bool
	peerTag  string
	logger   rua.Logger
	certFile string
	keyFile  string
	upgrader *websocket.Upgrader
}

func NewWebsocketListener(addr string, gs rua.GameServer) *websocketListener {
	return &websocketListener{
		name:     "WebsocketListener",
		addr:     addr,
		path:     "/",
		gs:       gs,
		guardian: nil,
		peerTag:  "websocket",
		logger:   rua.DefaultLogger(),
		certFile: "",
		keyFile:  "",
		upgrader: &websocket.Upgrader{},
	}
}

func (l *websocketListener) WithName(n string) *websocketListener {
	l.name = n
	return l
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

func (l *websocketListener) WithGuardian(g func(w http.ResponseWriter, r *http.Request) bool) *websocketListener {
	l.guardian = g
	return l
}

func (l *websocketListener) Start() error {
	http.HandleFunc(l.path, func(w http.ResponseWriter, r *http.Request) {
		if l.guardian == nil || l.guardian(w, r) {
			// upgrade http to websocket
			c, err := l.upgrader.Upgrade(w, r, nil)
			if err != nil {
				l.logger.Error("rua.WebsocketListener.Upgrade:", err)
				return
			}

			l.gs.AddPeer(NewWebsocketPeer(c, l.gs).WithLogger(l.logger).WithTag(l.peerTag))
		}
	})
	l.logger.Infof("%s is listening at %s", l.name, l.addr)

	if len(l.certFile) != 0 && len(l.keyFile) != 0 {
		return http.ListenAndServeTLS(l.addr, l.certFile, l.keyFile, nil)
	} else {
		return http.ListenAndServe(l.addr, nil)
	}
}
