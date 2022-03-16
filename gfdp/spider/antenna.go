package spider

import (
	"context"

	"github.com/gametaverse/gfdp/db"
)

type Transaction struct {
	timestamp uint64
	block     uint64
	index     uint16
	raw       interface{}
}

type Antennaer interface {
	Init(rpc string, chainID int) error
	GetBlockHeight(context.Context) (uint64, error)
	GetTrxByNum(context.Context, uint64) ([]*Transaction, error)
	DealTrx(rawtrx *Transaction) (txes []*db.Transaction, err error)
}
