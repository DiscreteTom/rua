package websocket

import (
	"log"
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
}

func NewWebsocketListener(addr string, gs rua.GameServer) *websocketListener {
	return &websocketListener{
		addr:     addr,
		path:     "/",
		gs:       gs,
		guardian: nil,
		peerTag:  "websocket",
	}
}

func (l *websocketListener) WithPath(p string) *websocketListener {
	l.path = p
	return l
}

func (l *websocketListener) WithPeerTag(t string) *websocketListener {
	l.peerTag = t
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
				log.Print("upgrade:", err)
				return
			}

			l.gs.AddPeer(NewWebsocketPeer(c, l.gs).WithTag(l.peerTag))
		}
	})
	log.Println("websocket server is listening at", l.addr)
	return http.ListenAndServe(l.addr, nil)
}
