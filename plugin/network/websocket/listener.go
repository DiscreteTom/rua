package websocket

import (
	"DiscreteTom/rua"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type websocketListener struct {
	addr     string
	gs       rua.GameServer
	guardian func(w http.ResponseWriter, r *http.Request, gs *rua.GameServer) bool
}

func NewWebsocketListener(addr string, gs rua.GameServer) *websocketListener {
	return &websocketListener{
		addr:     addr,
		gs:       gs,
		guardian: nil,
	}
}

func (l *websocketListener) WithGuardian(g func(w http.ResponseWriter, r *http.Request, gs *rua.GameServer) bool) *websocketListener {
	l.guardian = g
	return l
}

func (l *websocketListener) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if l.guardian == nil || l.guardian(w, r, &l.gs) {
			handler(w, r, l.gs)
		}
	})
	log.Println("websocket server is listening at", l.addr)
	return http.ListenAndServe(l.addr, nil)
}

func handler(w http.ResponseWriter, r *http.Request, gs rua.GameServer) {
	// upgrade http to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	p := &websocketPeer{c: c, gs: gs}
	gs.AddPeer(p)
}
