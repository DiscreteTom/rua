package rua

func WriteOrLog(p Peer, msg []byte) {
	if err := p.Write(msg); err != nil {
		p.Logger().Errorf("rua.WriteOrLog.[id=%d,tag=%s]: %s", p.Id(), p.Tag(), err)
	}
}
