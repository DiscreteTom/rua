package websocket

import (
	"log"
	"net/http"

	"github.com/DiscreteTom/rua"
)

type websocketCascadeLeader struct {
	addr string
	path string
	gs   rua.GameServer
}

func NewWebsocketCascadeLeader(addr string, gs rua.GameServer) *websocketCascadeLeader {
	return &websocketCascadeLeader{
		addr: addr,
		path: "/",
		gs:   gs,
	}
}

func (l *websocketCascadeLeader) WithPath(p string) *websocketCascadeLeader {
	l.path = p
	return l
}

func (l *websocketCascadeLeader) Start() error {
	http.HandleFunc(l.path, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		l.gs.AddPeer(NewWebsocketPeer(c, l.gs))
	})
	log.Println("websocket server is listening at", l.addr)
	return http.ListenAndServe(l.addr, nil)
}
