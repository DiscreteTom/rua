package kcp

import (
	"log"
	"net"

	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/peers/network"

	"github.com/xtaci/kcp-go/v5"
)

type KcpListener struct {
	addr         string
	gs           rua.GameServer
	key          []byte
	bufSize      int
	dataShards   int
	parityShards int
	crypt        string
	peerTimeout  int // in ms
	guardian     func(c *kcp.UDPSession, gs rua.GameServer) bool
	peerTag      string
	logger       rua.Logger
	maxAttempts  int
	upgrader     func(c *kcp.UDPSession) (net.Conn, error)
}

func NewKcpListener(addr string, gs rua.GameServer, key []byte, bufSize int) *KcpListener {
	return &KcpListener{
		addr:         addr,
		gs:           gs,
		key:          key,
		bufSize:      bufSize,
		dataShards:   10,
		parityShards: 3,
		crypt:        "aes",
		peerTimeout:  1000,
		guardian:     nil,
		peerTag:      "kcp",
		logger:       rua.GetDefaultLogger(),
		maxAttempts:  10,
		upgrader:     func(c *kcp.UDPSession) (net.Conn, error) { return c, nil },
	}
}

func (l *KcpListener) WithLogger(logger rua.Logger) *KcpListener {
	l.logger = logger
	return l
}

func (l *KcpListener) WithPeerTag(t string) *KcpListener {
	l.peerTag = t
	return l
}

func (l *KcpListener) WithDataShards(shards int) *KcpListener {
	l.dataShards = shards
	return l
}

func (l *KcpListener) WithParityShards(shards int) *KcpListener {
	l.parityShards = shards
	return l
}

func (l *KcpListener) WithCrypt(crypt string) *KcpListener {
	l.crypt = crypt
	return l
}

func (l *KcpListener) WithPeerTimeout(t int) *KcpListener {
	l.peerTimeout = t
	return l
}

func (l *KcpListener) WithGuardian(g func(c *kcp.UDPSession, gs rua.GameServer) bool) *KcpListener {
	l.guardian = g
	return l
}

func (l *KcpListener) WithMaxAttempts(count int) *KcpListener {
	l.maxAttempts = count
	return l
}

func (l *KcpListener) WithUpgrader(f func(c *kcp.UDPSession) (net.Conn, error)) *KcpListener {
	l.upgrader = f
	return l
}

func (l *KcpListener) Start() error {
	l.logger.Infof("kcp listener is listening at %s", l.addr)
	block, err := blockCrypt(l.crypt, l.key)
	if err != nil {
		return err
	}
	listener, err := kcp.ListenWithOptions(l.addr, block, l.dataShards, l.parityShards)
	if err != nil {
		return err
	}

	attempts := 0
	for {
		c, err := listener.AcceptKCP()
		if err != nil {
			if attempts < l.maxAttempts {
				l.logger.Error(err)
				attempts++
			} else {
				return err
			}
		} else { // err == nil
			attempts = 0
			if l.guardian == nil || l.guardian(c, l.gs) {
				con, err := l.upgrader(c)
				if err != nil {
					l.logger.Error(err)
				} else {
					l.gs.AddPeer(
						network.NewNetPeer(con, l.gs, l.bufSize, l.peerTimeout).
							WithLogger(l.logger).
							WithTag(l.peerTag),
					)
				}
			}
		}
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
