package spider

import (
	"context"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/eoscanada/eos-go"

	"github.com/gametaverse/gfdp/db"
	"github.com/gametaverse/gfdp/utils"
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
	for i, trx := range blk.Transactions {
		trxes = append(trxes, &Transaction{
			timestamp: utils.StartSecForDay(uint64(blk.Timestamp.Unix())),
			raw:       &trx,
			block:     num,
			index:     uint16(i),
		})
	}
	return
}

func (ata *EOSAntenna) DealTrx(rawtrx *Transaction) (txes []*db.Transaction, err error) {
	trx, ok := rawtrx.raw.(*eos.TransactionReceipt)
	if !ok {
		err = ErrUnknownTrx
		return
	}
	if trx.TransactionReceiptHeader.Status != eos.TransactionStatusExecuted {
		err = nil
		return
	}
	if nil == trx.Transaction.Packed {
		err = nil
		return
	}
	strx, err := trx.Transaction.Packed.Unpack()
	if err != nil {
		log.Error("trx.Transaction.Packed.Unpack error:%s", err.Error())
		return
	}
	for i, act := range strx.Actions {
		cName := act.Account
		from := ""
		for j, auth := range act.Authorization {
			if auth.Permission.String() == "active" {
				from = auth.Actor.String()
				tx := &db.Transaction{
					Block:     rawtrx.block,
					Index:     (rawtrx.index << 8) | (uint16(i) << 4) | uint16(j),
					Timestamp: time.Unix(int64(rawtrx.timestamp), 0),
					From:      strings.ToLower(from),
					To:        strings.ToLower(cName.String()),
				}
				txes = append(txes, tx)
			}
		}
	}
	err = nil
	return
}
