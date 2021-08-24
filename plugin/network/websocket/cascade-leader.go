package websocket

import (
	"net/http"

	"github.com/DiscreteTom/rua"
)

type websocketCascadeLeader struct {
	addr    string
	path    string
	gs      rua.GameServer
	peerTag string
	logger  rua.Logger
}

func NewWebsocketCascadeLeader(addr string, gs rua.GameServer) *websocketCascadeLeader {
	return &websocketCascadeLeader{
		addr:    addr,
		path:    "/",
		gs:      gs,
		peerTag: "websocket/cascade/leader",
		logger:  rua.GetDefaultLogger(),
	}
}

func (l *websocketCascadeLeader) WithLogger(logger rua.Logger) *websocketCascadeLeader {
	l.logger = logger
	return l
}

func (l *websocketCascadeLeader) WithPeerTag(t string) *websocketCascadeLeader {
	l.peerTag = t
	return l
}

func (l *websocketCascadeLeader) WithPath(p string) *websocketCascadeLeader {
	l.path = p
	return l
}

func (l *websocketCascadeLeader) Start() error {
	http.HandleFunc(l.path, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			l.logger.Error(err)
			return
		}
		l.gs.AddPeer(NewWebsocketPeer(c, l.gs).WithLogger(l.logger).WithTag(l.peerTag))
	})
	l.logger.Info("websocket cascade leader is listening at", l.addr)
	return http.ListenAndServe(l.addr, nil)
}
