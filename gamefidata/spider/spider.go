package spider

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/fourbyte"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Spider struct {
	ctx         context.Context
	cancelFun   context.CancelFunc
	interval    int
	chainID     int
	ethcli      *ethclient.Client
	dbClient    *mongo.Client
	db          *mongo.Database
	games       []*Game
	topBlock    uint64
	curBlock    uint64
	bottomBlock uint64
	mongoURI    string
	rpcAddr     string
	backward    bool
	monitorTbl  *mongo.Collection
}

func (sp *Spider) Init() (err error) {
	log.Info("init")
	sp.ctx, sp.cancelFun = context.WithCancel(context.Background())

	sp.ethcli, err = ethclient.Dial(sp.rpcAddr)
	if err != nil {
		log.Error("Dial error:%s", err.Error())
		return err
	}

	err = sp.initDB(sp.mongoURI)
	if err != nil {
		log.Error("Init mongon error:%s", err.Error())
		return err
	}

	return
}

func (sp *Spider) initDB(URI string) (err error) {
	sp.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(sp.ctx, 10*time.Second)
	err = sp.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = sp.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	sp.db = sp.dbClient.Database(db.DBName)
	if sp.db == nil {
		log.Error("db solana-spl is null, please init db first")
		return
	}

	sp.monitorTbl = sp.db.Collection(db.MonitorTableName)
	if sp.monitorTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}

	return
}

func (sp *Spider) Run() {
	log.Info("Spider Run")
	if sp.backward {
		sp.goBackward()
	} else {
		sp.goForward()
	}
}

func (sp *Spider) loadTopBlock() (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{
		"_id": db.MonitorFieldName,
	}
	curs, err := sp.monitorTbl.Find(ctx, filter)
	if err != nil {
		log.Error("Find monitor error:", err.Error())
		return err
	}

	for curs.Next(ctx) {
		var m db.Monitor
		curs.Decode(&m)
		log.Info("t_monitor:%v", m)
		sp.topBlock = m.TopBlock
		break
	}

	if sp.topBlock == 0 {
		// No such record
		log.Info("curs:%v current %v", curs, curs.Current)
		sp.topBlock, err = sp.getBlockHeight()
		if err != nil {
			log.Error("get block height error:", err.Error())
			return err
		}
	}
	log.Info("topBlock:%v", sp.topBlock)
	return
}

func (sp *Spider) getBlockHeight() (height uint64, err error) {
	height, err = sp.ethcli.BlockNumber(sp.ctx)
	if err != nil {
		log.Error("get block height error:%s", err)
		return
	}
	return
}

func (sp *Spider) goForward() {

	err := sp.loadTopBlock()
	if err != nil {
		log.Error("load top block:", err.Error())
		return
	}
	sp.curBlock = sp.topBlock
	for {
		err := sp.dealBlock(sp.curBlock)
		if err != nil {
			log.Error("deal block[%d]:%s", sp.curBlock, err.Error())
			time.Sleep(time.Second * time.Duration(sp.interval))
			continue
		}
		if sp.curBlock%10 == 0 {
			sp.storeTopBlock(sp.curBlock)
		}
		sp.curBlock += 1
	}
}

func (sp *Spider) storeTopBlock(number uint64) (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"topblock": number,
		},
	}
	_, err = sp.monitorTbl.UpdateByID(ctx, db.MonitorFieldName, update, opt)
	if err != nil {
		log.Error("Update top block error: ", err.Error())
		return
	}
	log.Info("Update top block:%d ", number)
	return
}

func (sp *Spider) dealGame(game *Game, blk *types.Block, trx *types.Transaction) (err error) {

	gameTbl := sp.db.Collection("t_" + game.info.ID)
	if gameTbl == nil {
		log.Error("create collection for game[%s] error:%s", game.info.ID, err.Error())
		return
	}

	for _, c := range game.info.Contracts {
		msg, err := trx.AsMessage(types.NewEIP155Signer(big.NewInt(int64(sp.chainID))), big.NewInt(0))
		if err != nil {
			return nil // success when can not AsMessage
		}
		if trx.To() == nil {
			return nil // done with 0x0000...000
		}
		if strings.EqualFold(c.Address, trx.To().Hex()) {
			log.Info("[%s] %s send transaction to contract:%v %v:%v", trx.Hash().Hex(), msg.From().Hex(), trx.To().Hex(), blk.Header().Time, blk.NumberU64())
			inputData := fmt.Sprintf("0x%x", trx.Data())
			method := ""
			if len(inputData) > 10 {
				methodTxt := inputData[:10]

				method, err = fourbyte.DB.Get(methodTxt)
				if err != nil {
					// TODO: retry
					log.Error("Get fourbyte error:%s", err.Error())
				}
			}

			trxModel := &db.Transaction{
				GameID:    game.info.ID,
				Timestamp: blk.Header().Time,
				ID:        trx.Hash().Hex(),
				From:      msg.From().Hex(),
				To:        trx.To().Hex(),
				BlockNum:  blk.NumberU64(),
				Method:    method,
			}
			log.Info("trxModel:%v", trxModel)

			ctx, cancel := context.WithTimeout(sp.ctx, 3*time.Second)
			defer cancel()

			opt := mngopts.FindOneAndReplace()
			opt.SetUpsert(true)

			filter := bson.M{
				"_id": trxModel.ID,
			}

			rst := gameTbl.FindOneAndReplace(ctx, filter, trxModel, opt)
			if rst.Err() != nil {
				if !strings.Contains(rst.Err().Error(), "no documents in result") {
					log.Error("update transaction[%s] error: %s", trxModel.ID, rst.Err().Error())
					return err // return to show error
				}
			}
		}
	}
	return nil
}

func (sp *Spider) dealBlock(number uint64) (err error) {
	blk, err := sp.ethcli.BlockByNumber(sp.ctx, big.NewInt(int64(number)))
	if err != nil {
		//log.Error("get block[%d] error:%s", number, err.Error())
		if strings.Contains(err.Error(), "block header indicates no transactions") {
			return nil
		}
		return
	}
	for _, trx := range blk.Transactions() {
		for _, g := range sp.games {
			err = sp.dealGame(g, blk, trx)
			if err != nil {
				log.Error("deal game error: %s", err.Error())
				return err // for show errors
			}
		}
	}
	return nil
}

func (sp *Spider) goBackward() {

}
