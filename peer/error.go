package peer

import "errors"

var (
	ErrClosed = errors.New("peer already closed")
)
