package main

import (
	"crypto/sha1"
	"log"
	"math/rand"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	sess, err := kcp.DialWithOptions("localhost:8081", block, 10, 3)
	if err != nil {
		log.Fatal(err)
	}

	// write a byte to start
	if _, err := sess.Write([]byte{0}); err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, 8)
		// get server's echo request
		if _, err := sess.Read(buf); err != nil {
			log.Fatal(err)
		}
		// echo back
		// you can wait for a random period to simulate a latency
		time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
		if _, err := sess.Write(buf); err != nil {
			log.Fatal(err)
		}
	}
}
