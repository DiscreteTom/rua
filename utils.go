package rua

import "time"

func Wait(ms uint64, c chan<- bool) {
	time.Sleep(time.Millisecond * time.Duration(ms))
	c <- true
}
