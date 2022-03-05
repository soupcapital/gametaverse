package spider

import (
	"context"

	"github.com/gametaverse/gamefidata/db"
)

type Transaction struct {
	timestamp uint64
	raw       interface{}
	block     uint64
}

type Antennaer interface {
	Init(rpc string, chainID int) error
	GetBlockHeight(context.Context) (uint64, error)
	DealTrx4Game(game *GameInfo, trx *Transaction) ([]*db.Action, error)
	GetTrxByNum(context.Context, uint64) ([]*Transaction, error)
}
