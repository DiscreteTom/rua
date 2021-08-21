package main

import (
	"DiscreteTom/rua"
	"DiscreteTom/rua/plugin/network/websocket"
	"fmt"
	"log"
)

func main() {
	errChan := make(chan error)
	s := rua.NewFifoServer().SetHandleKeyboardInterrupt(true)

	go func() {
		errChan <- websocket.NewWebsocketListener(":8080", s).Start()
	}()

	serverErrsChan := make(chan []error)
	go func() {
		serverErrsChan <- s.Start(broadcastFifoHandler)
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

func broadcastFifoHandler(peers map[int]rua.Peer, msg *rua.PeerMsg, _ *rua.FifoServer) (errs []error) {
	// compact msg in one byte array
	result := []byte{}
	result = append(result, []byte(fmt.Sprintf("from %d:\n", msg.PeerId))...)
	result = append(result, msg.Data...)
	result = append(result, '\n')
	// broadcast to everyone
	for _, p := range peers {
		go p.Write(result)
	}
	return
}
