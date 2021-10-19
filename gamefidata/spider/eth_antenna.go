package spider

import (
	"context"
	"math/big"
	"strings"

	"github.com/cz-theng/czkit-go/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/utils"
)

type ETHAntenna struct {
	ethcli  *ethclient.Client
	chainID int
}

func NewETHAntenna() *ETHAntenna {
	antenna := &ETHAntenna{}
	return antenna
}

func (ata *ETHAntenna) Init(rpc string, chainID int) (err error) {
	log.Info("ETHAntenna init")
	ata.chainID = chainID
	ata.ethcli, err = ethclient.Dial(rpc)
	if err != nil {
		log.Error("Dial error:%s", err.Error())
		return err
	}
	return
}

func (ata *ETHAntenna) GetBlockHeight(ctx context.Context) (height uint64, err error) {
	height, err = ata.ethcli.BlockNumber(ctx)
	if err != nil {
		return
	}
	return
}

func (ata *ETHAntenna) GetTrxByNum(ctx context.Context, num uint64) (trxes []*Transaction, err error) {
	blk, err := ata.ethcli.BlockByNumber(ctx, big.NewInt(int64(num)))
	if err != nil {
		return nil, err
	}
	for _, trx := range blk.Transactions() {
		trxes = append(trxes, &Transaction{
			timestamp: utils.StartSecForDay(blk.Header().Time),
			raw:       trx,
		})
	}
	return

}

func (ata *ETHAntenna) DealTrx4Game(game *Game, rawtrx *Transaction) (actions []*db.Action, err error) {
	trx, ok := rawtrx.raw.(*types.Transaction)
	if !ok {
		return nil, ErrUnknownTrx
	}

	for _, c := range game.info.Contracts {
		msg, err := trx.AsMessage(types.NewLondonSigner(big.NewInt(int64(ata.chainID))), big.NewInt(0))
		if err != nil {
			log.Error("[%s:%v]AsMessage error:%s", trx.Hash().Hex(), trx.Type(), err.Error())
			continue
		}
		if trx.To() == nil {
			continue // done with 0x0000...000
		}
		if strings.EqualFold(c.Address, trx.To().Hex()) {
			from := msg.From().Hex()
			action := &db.Action{
				GameID:    game.info.ID,
				Timestamp: rawtrx.timestamp,
				User:      from,
				Count:     1,
			}
			actions = append(actions, action)

		}
	}
	return actions, nil
}
