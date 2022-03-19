package rpc

import "errors"

var (
	ErrNoCahceItem  = errors.New("no such cache item")
	ErrUnknownChain = errors.New("unknown chain")
)
