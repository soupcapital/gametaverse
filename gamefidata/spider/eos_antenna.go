package spider

import (
	"context"

	"github.com/cz-theng/czkit-go/log"
	"github.com/eoscanada/eos-go"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/utils"
)

type EOSAntenna struct {
	eoscli  *eos.API
	chainID int
}

func NewEOSAntenna() *EOSAntenna {
	antenna := &EOSAntenna{}
	return antenna
}

func (ata *EOSAntenna) Init(rpc string, chainID int) (err error) {
	log.Info("EOSAntenna init")
	ata.chainID = chainID
	ata.eoscli = eos.New(rpc)
	return
}

func (ata *EOSAntenna) GetBlockHeight(ctx context.Context) (height uint64, err error) {
	infoRsp, err := ata.eoscli.GetInfo(ctx)
	if err != nil {
		return
	}
	height = uint64(infoRsp.HeadBlockNum)
	return
}

func (ata *EOSAntenna) GetTrxByNum(ctx context.Context, num uint64) (trxes []*Transaction, err error) {
	blk, err := ata.eoscli.GetBlockByNum(ctx, uint32(num))
	if err != nil {
		return nil, err
	}
	for _, trx := range blk.Transactions {
		trxes = append(trxes, &Transaction{
			timestamp: utils.StartSecForDay(uint64(blk.Timestamp.Unix())),
			raw:       &trx,
		})
	}
	return
}

func (ata *EOSAntenna) DealTrx4Game(game *Game, rawtrx *Transaction) (actions []*db.Action, err error) {
	trx, ok := rawtrx.raw.(*eos.TransactionReceipt)
	if !ok {
		return nil, ErrUnknownTrx
	}
	for _, c := range game.info.Contracts {
		if trx.TransactionReceiptHeader.Status != eos.TransactionStatusExecuted {
			continue
		}
		if nil == trx.Transaction.Packed {
			continue
		}
		strx, err := trx.Transaction.Packed.Unpack()
		if err != nil {
			log.Error("trx.Transaction.Packed.Unpack error:%s", err.Error())
			continue
		}
		for _, act := range strx.Actions {
			cName := act.Account
			if c.Address != cName.String() {
				continue
			}
			from := ""
			for _, auth := range act.Authorization {
				if auth.Permission.String() == "active" {
					from = auth.Actor.String()
					action := &db.Action{
						GameID:    game.info.ID,
						Timestamp: rawtrx.timestamp,
						User:      from,
						Count:     1,
					}
					actions = append(actions, action)
				}
			}
		}
	}
	return
}
