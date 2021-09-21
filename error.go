package rua

import "errors"

var (
	ErrPeerClosed       = errors.New("peer already closed")
	ErrPeerNotExist     = errors.New("peer not exist")
	ErrPeerWriteTimeout = errors.New("peer write timeout")
)
