package kcp

import (
	"DiscreteTom/rua/pkg/model"
	"log"

	"github.com/xtaci/kcp-go/v5"
)

type kcpListener struct {
	addr         string
	gs           model.GameServer
	key          []byte
	bufSize      int
	dataShards   int
	parityShards int
	crypt        string
}

func NewKcpListener(addr string, gs model.GameServer, key []byte, bufSize int) *kcpListener {
	return &kcpListener{
		addr:         addr,
		gs:           gs,
		key:          key,
		bufSize:      bufSize,
		dataShards:   10,
		parityShards: 3,
		crypt:        "aes",
	}
}

func (l *kcpListener) WithDataShards(shards int) *kcpListener {
	l.dataShards = shards
	return l
}

func (l *kcpListener) WithParityShards(shards int) *kcpListener {
	l.parityShards = shards
	return l
}

func (l *kcpListener) WithCrypt(crypt string) *kcpListener {
	l.crypt = crypt
	return l
}

func (l *kcpListener) Start() error {
	log.Println("kcp server is listening at", l.addr)
	block, _ := blockCrypt(l.crypt, l.key)
	listener, err := kcp.ListenWithOptions(l.addr, block, l.dataShards, l.parityShards)
	if err != nil {
		return err
	}
	for {
		c, err := listener.AcceptKCP()
		if err != nil {
			return err
		}
		p := &kcpPeer{c: c, gs: l.gs, bufSize: l.bufSize}
		l.gs.AddPeer(p)
	}
}

func blockCrypt(crypt string, key []byte) (kcp.BlockCrypt, error) {
	switch crypt {
	case "sm4":
		return kcp.NewSM4BlockCrypt(key[:16])
	case "tea":
		return kcp.NewTEABlockCrypt(key[:16])
	case "xor":
		return kcp.NewSimpleXORBlockCrypt(key)
	case "none":
		return kcp.NewNoneBlockCrypt(key)
	case "aes":
		return kcp.NewAESBlockCrypt(key)
	case "aes-128":
		return kcp.NewAESBlockCrypt(key[:16])
	case "aes-192":
		return kcp.NewAESBlockCrypt(key[:24])
	case "blowfish":
		return kcp.NewBlowfishBlockCrypt(key)
	case "twofish":
		return kcp.NewTwofishBlockCrypt(key)
	case "cast5":
		return kcp.NewCast5BlockCrypt(key[:16])
	case "3des":
		return kcp.NewTripleDESBlockCrypt(key[:24])
	case "xtea":
		return kcp.NewXTEABlockCrypt(key[:16])
	case "salsa20":
		return kcp.NewSalsa20BlockCrypt(key)
	default:
		log.Println("unknown cryption, use aes")
		return kcp.NewAESBlockCrypt(key)
	}
}
