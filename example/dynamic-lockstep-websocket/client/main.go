package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		// got server's echo request
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		// echo back
		// you can wait for a random period to simulate a latency
		time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
		c.WriteMessage(websocket.TextMessage, message)
	}
}
