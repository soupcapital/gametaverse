package eth

import (
	"context"
	"math/big"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Game struct {
	ctx       context.Context
	cancelFun context.CancelFunc
	info      *GameInfo
	ethcli    *ethclient.Client
}

func NewGame(info *GameInfo) *Game {
	gm := &Game{
		info: info,
	}
	return gm
}

func (gm *Game) Init() (err error) {
	log.Info("init")
	gm.ctx, gm.cancelFun = context.WithCancel(context.Background())

	gm.ethcli, err = ethclient.Dial(gm.info.RPCAddr)
	if err != nil {
		log.Error("Dial error:%s", err.Error())
		return err
	}
	return
}

func (gm *Game) Run() (err error) {
	number := big.NewInt(0)
	for _, c := range gm.info.Contracts {
		num := big.NewInt(0)
		num.SetString(c.StartBlock, 10)
		if num.Cmp(number) < 0 ||
			(0 == big.NewInt(0).Cmp(number)) {
			number = num
		}
	}
	log.Info("Deal for game:%s", gm.info.Name)
	for {
		blk, err := gm.ethcli.BlockByNumber(gm.ctx, number)
		if err != nil {
			log.Error("get block error:%s", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		for _, trx := range blk.Transactions() {
			for _, c := range gm.info.Contracts {
				msg, err := trx.AsMessage(types.NewEIP155Signer(big.NewInt(int64(gm.info.ChainID))), big.NewInt(0))
				if err != nil {
					//log.Error("AsMessage error:%s", err.Error())
					continue
				}
				//log.Info("msg from:%v to:%v", msg.From(), trx.To())
				if trx.To() == nil {
					continue
				}
				if c.Address == trx.To().Hex() {
					log.Info("[%s] %s send transaction to contract:%v", trx.Hash().Hex(), msg.From().Hex(), trx.To().Hex())
				}
			}
		}
		//log.Info("Done for block:%v", blk.Number)
		number.Add(number, big.NewInt(1))
	}
	return
}
