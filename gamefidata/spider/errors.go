package spider

import (
	"errors"
)

var (
	ErrAsMessage  = errors.New("AsMessage error")
	ErrGetBlock   = errors.New("Get block  error")
	ErrUnknownTrx = errors.New("Unknown transaction for this chain")
)
