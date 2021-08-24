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

	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		BeforeAddPeer(func(newPeer rua.Peer, peers map[int]rua.Peer, s *rua.EventDrivenServer) {
			// notify existing players
			broadcastSync(peers, []byte(fmt.Sprintf(
				"Player[%d] added. Current player count: %d",
				newPeer.GetId(), s.GetPeerCount()+1,
			)))
		}).
		AfterAddPeer(func(newPeer rua.Peer, peers map[int]rua.Peer, s *rua.EventDrivenServer) {
			// notify the new Peer
			newPeer.Write([]byte(fmt.Sprintf("Current player count: %d", s.GetPeerCount())))
			// if all players are arrived, start the game and notify all players
			if s.GetPeerCount() == playerCount {
				game.Started = true
				broadcastSync(peers, []byte("Game started"))
			}
		}).
		AfterRemovePeer(func(targetId int, peers map[int]rua.Peer, s *rua.EventDrivenServer) {
			if !game.Over { // players are not removed by the server
				if s.GetPeerCount() == 0 {
					// no player left, end the game
					s.Stop()
				} else {
					// notify remaining players
					broadcastSync(peers, []byte(fmt.Sprintf(
						"Player[%d] leaved. Current player count: %d",
						targetId, s.GetPeerCount(),
					)))
					// not enough player, stop the game to wait for enough players
					if game.Started {
						game.Started = false
						broadcastSync(peers, []byte("Game stopped"))
					}
				}
			}
		}).
		OnPeerMsg(func(peers map[int]rua.Peer, msg *rua.PeerMsg, s *rua.EventDrivenServer) {
			if game.Started {
				processGameLogic(peers, msg, s, game)
			} else {
				// game not started, echo an error
				peers[msg.PeerId].Write([]byte("Game has not been started"))
			}
		})

	errChan := make(chan error)
	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).
			WithGuardian(func(_ http.ResponseWriter, _ *http.Request, gs rua.GameServer) bool {
				return gs.GetPeerCount() < playerCount // only playerCount players can be accepted
			}).
			Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start()
	}()

	select {
	case err := <-errChan:
		s.GetLogger().Error(err)
	case errs := <-serverErrsChan:
		if len(errs) != 0 {
			s.GetLogger().Error(errs)
		}
		break
	}
}

func processGameLogic(peers map[int]rua.Peer, msg *rua.PeerMsg, gs rua.GameServer, state *Game) {
	// dead player can not attack
	if state.PlayerHealth[msg.PeerId] == 0 {
		go peers[msg.PeerId].Write([]byte("You are dead and can not attack"))
		return
	}

	// attack take effects
	for i, p := range peers {
		if i != msg.PeerId { // not the attacker
			if state.PlayerHealth[i] > 0 { // alive
				state.PlayerHealth[i]--
				if state.PlayerHealth[i] != 0 {
					go p.Write([]byte(fmt.Sprintf(
						"Got attack from player[%d], current health: %d",
						msg.PeerId, state.PlayerHealth[i],
					)))
				} else {
					go p.Write([]byte(fmt.Sprintf(
						"Got attack from player[%d], you are dead.\n Entering watcher mode.",
						msg.PeerId,
					)))
					state.AlivePlayerCount--
				}
			} else { // dead player in watcher mode
				go p.Write([]byte(fmt.Sprintf(
					"Player[%d] initiated attack",
					msg.PeerId,
				)))
			}
		}
	}

	// game end?
	if state.AlivePlayerCount == 1 { // the attacker won
		broadcastSync(peers, []byte(fmt.Sprintf(
			"Game Over.\nPlayer[%d] won!\n",
			msg.PeerId,
		)))
		state.Over = true
		gs.Stop()
	}
}

func broadcastSync(peers map[int]rua.Peer, msg []byte) {
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(p rua.Peer) {
			p.Write(msg)
			wg.Done()
		}(p)
	}
	wg.Wait()
}
