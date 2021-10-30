package main

import (
	"fmt"
	"os"

	"github.com/DiscreteTom/rua"
)

func main() {
	file, err := rua.DefaultFileNode().Filename("log.txt").Go()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stdio := rua.DefaultStdioNode().OnMsg(func(b []byte) {
		file.Write(b)
	}).Go()

	rua.NewCtrlc().OnSignal(func() {
		file.Stop()
		stdio.Stop()
	}).Wait()
}
