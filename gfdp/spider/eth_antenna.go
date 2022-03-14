package spider

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gametaverse/gfdp/db"
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
		if strings.Contains(err.Error(), "non-empty transaction list but block header indicates no transactions") {
			return nil, nil
		}
		return nil, err
	}
	for i, trx := range blk.Transactions() {
		trxes = append(trxes, &Transaction{
			timestamp: blk.Header().Time,
			block:     num,
			index:     uint16(i),
			raw:       trx,
		})
	}
	return
}

func (ata *ETHAntenna) DealTrx(rawtrx *Transaction) (txes []*db.Transaction, err error) {

	trx, ok := rawtrx.raw.(*types.Transaction)
	if !ok {
		return nil, ErrUnknownTrx
	}
	if trx.To() == nil {
		return // done with 0x0000...000
	}
	msg, err := trx.AsMessage(types.NewLondonSigner(big.NewInt(int64(ata.chainID))), big.NewInt(0))
	if err != nil {
		log.Error("[%s:%v]AsMessage error:%s", trx.Hash().Hex(), trx.Type(), err.Error())
		return
	}
	to := trx.To().Hex()
	from := msg.From().Hex()
	tx := &db.Transaction{
		Block:     rawtrx.block,
		Index:     rawtrx.index,
		Timestamp: time.Unix(int64(rawtrx.timestamp), 0),
		From:      strings.ToLower(from),
		To:        strings.ToLower(to),
	}
	txes = append(txes, tx)

	return
}
