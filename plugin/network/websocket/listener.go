package websocket

import (
	"DiscreteTom/rua/pkg/model"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type websocketListener struct {
	addr string
	gs   model.GameServer
}

func NewWebsocketListener(addr string, gs model.GameServer) *websocketListener {
	return &websocketListener{
		addr: addr,
		gs:   gs,
	}
}

func (l *websocketListener) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, l.gs)
	})
	log.Println("websocket server is listening at", l.addr)
	return http.ListenAndServe(l.addr, nil)
}

func handler(w http.ResponseWriter, r *http.Request, gs model.GameServer) {
	// upgrade http to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	p := &websocketPeer{c: c, gs: gs}
	gs.AddPeer(p)
}
