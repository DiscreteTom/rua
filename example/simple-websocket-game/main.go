package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

const playerCount = 5
const playerMaxHealth = 10

type Game struct {
	PlayerHealth map[int]int // health of each player
}

func main() {
	game := &Game{
		PlayerHealth: map[int]int{},
	}
	// each player has 10 health point
	for i := 0; i < playerCount; i++ {
		game.PlayerHealth[i] = playerMaxHealth
	}

	errChan := make(chan error)
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		On(rua.AppendPeerMsg, func(peers map[int]rua.Peer, msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			statefulEventDrivenHandler(peers, msg, s, game)
		})

	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).WithGuardian(func(_ http.ResponseWriter, _ *http.Request, gs rua.GameServer) bool {
			return gs.GetPeerCount() < playerCount // only playerCount players can be accepted
		}).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		log.Println(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			log.Println(errs)
		}
		break
	}
}

func statefulEventDrivenHandler(peers map[int]rua.Peer, msg *rua.PeerMsg, _ *rua.EventDrivenServer, state *Game) {
	if state.PlayerHealth[msg.PeerId] == 0 {
		// dead player can not attack
		go peers[msg.PeerId].Write([]byte("You are dead and can not attack\n"))
		return
	}

	for i, p := range peers {
		if i != msg.PeerId { // not the attacker
			if state.PlayerHealth[i] > 0 { // alive
				state.PlayerHealth[i]--
				if state.PlayerHealth[i] != 0 {
					go p.Write([]byte(fmt.Sprintf("Got attack from player[%d], current health: %d\n", msg.PeerId, state.PlayerHealth[i])))
				} else {
					go p.Write([]byte(fmt.Sprintf("Got attack from player[%d], you are dead.\n", msg.PeerId)))
				}
			}
		}
	}
}
