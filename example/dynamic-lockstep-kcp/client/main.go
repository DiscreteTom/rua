package main

import (
	"crypto/sha1"
	"encoding/binary"
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

	for {
		// write current time
		buf := make([]byte, 8)
		currentTime := time.Now().UnixMilli()
		binary.LittleEndian.PutUint64(buf, uint64(currentTime))
		// but wait a random period (no longer than 100 ms) to simulate a latency
		time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
		if _, err := sess.Write(buf); err != nil {
			log.Fatal(err)
		}
		// wait for next step
		if _, err := sess.Read(buf); err != nil {
			log.Fatal(err)
		}
	}
}
