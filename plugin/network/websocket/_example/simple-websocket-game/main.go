package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

const playerCount = 3
const playerMaxHealth = 10

type Game struct {
	PlayerHealth     map[int]int // health of each player
	Started          bool
	AlivePlayerCount int
	Over             bool
}

func main() {
	game := &Game{
		PlayerHealth:     map[int]int{},
		Started:          false,
		AlivePlayerCount: playerCount,
		Over:             false,
	}
	// init player's health
	for i := 0; i < playerCount; i++ {
		game.PlayerHealth[i] = playerMaxHealth
	}

	s := rua.NewEventDrivenServer()

	s.
		AfterAddPeer(func(newPeer rua.Peer) {
			// notify existing peers
			s.ForEachPeer(func(id int, peer rua.Peer) {
				if id != newPeer.Id() {
					peer.Write([]byte(fmt.Sprintf(
						"Player[%d] added. Current player count: %d",
						newPeer.Id(), s.PeerCount(),
					)))
				}
			})
			// notify the new Peer
			newPeer.Write([]byte(fmt.Sprintf("Current player count: %d", s.PeerCount())))
			// if all players are arrived, start the game and notify all players
			if s.PeerCount() == playerCount {
				game.Started = true
				broadcastSync(s, []byte("Game started"))
			}
		}).
		AfterRemovePeer(func(targetId int) {
			if !game.Over { // players are not removed by the server
				if s.PeerCount() == 0 {
					// no player left, end the game
					s.Stop()
				} else {
					// notify remaining players
					broadcastSync(s, []byte(fmt.Sprintf(
						"Player[%d] leaved. Current player count: %d",
						targetId, s.PeerCount(),
					)))
					// not enough player, stop the game to wait for enough players
					if game.Started {
						game.Started = false
						broadcastSync(s, []byte("Game stopped"))
					}
				}
			}
		}).
		OnPeerMsg(func(msg *rua.PeerMsg) {
			if game.Started {
				processGameLogic(msg, s, game)
			} else {
				// game not started, echo an error
				msg.Peer.Write([]byte("Game has not been started"))
			}
		})

	errChan := make(chan error)
	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).
			WithGuardian(func(_ http.ResponseWriter, _ *http.Request) bool {
				return s.PeerCount() < playerCount // only playerCount players can be accepted
			}).
			Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		s.Logger().Error(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			s.Logger().Error(errs)
		}
		break
	}
}

func processGameLogic(msg *rua.PeerMsg, gs *rua.EventDrivenServer, state *Game) {
	// dead player can not attack
	if state.PlayerHealth[msg.Peer.Id()] == 0 {
		go msg.Peer.Write([]byte("You are dead and can not attack"))
		return
	}

	// attack take effects
	gs.ForEachPeer(func(i int, p rua.Peer) {
		if i != msg.Peer.Id() { // not the attacker
			if state.PlayerHealth[i] > 0 { // alive
				state.PlayerHealth[i]--
				if state.PlayerHealth[i] != 0 {
					go p.Write([]byte(fmt.Sprintf(
						"Got attack from player[%d], current health: %d",
						msg.Peer.Id(), state.PlayerHealth[i],
					)))
				} else {
					go p.Write([]byte(fmt.Sprintf(
						"Got attack from player[%d], you are dead.\n Entering watcher mode.",
						msg.Peer.Id(),
					)))
					state.AlivePlayerCount--
				}
			} else { // dead player in watcher mode
				go p.Write([]byte(fmt.Sprintf(
					"Player[%d] initiated attack",
					msg.Peer.Id(),
				)))
			}
		}
	})

	// game end?
	if state.AlivePlayerCount == 1 { // the attacker won
		broadcastSync(gs, []byte(fmt.Sprintf(
			"Game Over.\nPlayer[%d] won!\n",
			msg.Peer.Id(),
		)))
		state.Over = true
		gs.Stop()
	}
}

func broadcastSync(s *rua.EventDrivenServer, msg []byte) {
	var wg sync.WaitGroup
	s.ForEachPeer(func(id int, peer rua.Peer) {
		wg.Add(1)
		go func() {
			peer.Write(msg)
			wg.Done()
		}()
	})
	wg.Wait()
}
