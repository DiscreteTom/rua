package main

import "github.com/DiscreteTom/rua"

func main() {
	bc := rua.NewBroadcaster()

	// start tcp listener
	tcp, _ := rua.NewTcpListener("127.0.0.1:8080").OnNewPeer(func(tn *rua.TcpNode) {
		// new peer will be added to the broadcaster
		bc.AddTarget(tn.OnInput(func(b []byte) {
			// new message will be sent to the broadcaster
			bc.Write(b)
		}).Go())
	}).Go()

	// also print to stdout
	bc.AddTarget(rua.DefaultStdioNode().Go())

	rua.NewCtrlc().OnSignal(func() {
		bc.StopAll()
		tcp.Stop()
	}).Wait()
}
