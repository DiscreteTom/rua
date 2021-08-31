package main

import (
	"crypto/sha1"
	"log"
	"math/rand"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	c, err := kcp.DialWithOptions("localhost:8081", block, 10, 3)
	if err != nil {
		log.Fatal(err)
	}

	// smux client
	session, err := smux.Client(c, nil)
	if err != nil {
		panic(err)
	}

	// create a stream
	con, err := session.OpenStream()
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 8)
		// get server's echo request
		if _, err := con.Read(buf); err != nil {
			log.Fatal(err)
		}
		// echo back
		// you can wait for a random period to simulate a latency
		time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
		if _, err := con.Write(buf); err != nil {
			log.Fatal(err)
		}
	}
}
