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

type ETHSpider struct {
	ctx              context.Context
	cancelFun        context.CancelFunc
	forwardInterval  float32
	backwardInterval float32
	chainID          int
	chain            string
	ethcli           *ethclient.Client
	dbClient         *mongo.Client
	db               *mongo.Database
	games            []*Game
	topBlock         uint64
	headBlock        uint64
	tailBlock        uint64
	bottomBlock      uint64
	mongoURI         string
	rpcAddr          string
	backward         bool
	monitorField     string
	monitorTbl       *mongo.Collection

	dauTbl   *mongo.Collection
	countTbl *mongo.Collection
}

func (sp *ETHSpider) Init() (err error) {
	log.Info("init")
	sp.ctx, sp.cancelFun = context.WithCancel(context.Background())

	sp.ethcli, err = ethclient.Dial(sp.rpcAddr)
	if err != nil {
		log.Error("Dial error:%s", err.Error())
		return err
	}

	sp.monitorField = db.MonitorFieldName + "_" + sp.chain
	err = sp.initDB(sp.mongoURI)
	if err != nil {
		log.Error("Init mongon error:%s", err.Error())
		return err
	}

	return
}

func (sp *ETHSpider) initDB(URI string) (err error) {
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

	sp.dauTbl = sp.db.Collection(db.DAUTableName)
	if sp.dauTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}

	sp.countTbl = sp.db.Collection(db.CountTableName)
	if sp.countTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}

	return
}

func (sp *ETHSpider) Run() {
	err := sp.loadTopBlock()
	if err != nil {
		log.Error("load top block:", err.Error())
		return
	}

	if sp.backward {
		sp.goBackward()
	} else {
		sp.goForward()
	}
}

func (sp *ETHSpider) loadTopBlock() (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{
		"_id": sp.monitorField,
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
		sp.topBlock, err = sp.getBlockHeight()
		if err != nil {
			log.Error("get block height error:", err.Error())
			return err
		}
	}
	log.Info("topBlock:%v", sp.topBlock)
	return
}

func (sp *ETHSpider) getBlockHeight() (height uint64, err error) {
	height, err = sp.ethcli.BlockNumber(sp.ctx)
	if err != nil {
		log.Error("get block height error:%s", err)
		return
	}
	log.Info("block height:%v", height)
	return
}

func (sp *ETHSpider) goForward() {
	log.Info("go forward")
	sp.headBlock = sp.topBlock
	for {
		err := sp.dealBlock(sp.headBlock)
		if err != nil {
			log.Error("deal block[%d]:%s", sp.headBlock, err.Error())
			time.Sleep(time.Duration(sp.forwardInterval * float32(time.Second)))
			continue
		}
		if sp.headBlock%10 == 0 {
			sp.storeTopBlock(sp.headBlock)
		}
		sp.headBlock += 1
	}
}

func (sp *ETHSpider) storeTopBlock(number uint64) (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"topblock": number,
		},
	}
	_, err = sp.monitorTbl.UpdateByID(ctx, sp.monitorField, update, opt)
	if err != nil {
		log.Error("Update top block error: ", err.Error())
		return
	}
	log.Info("Update top block:%d ", number)
	return
}

func (sp *ETHSpider) dealGame(game *Game, blk *types.Block, trx *types.Transaction) (err error) {

	gameTbl := sp.db.Collection("t_" + game.info.ID)
	if gameTbl == nil {
		log.Error("create collection for game[%s] error:%s", game.info.ID, err.Error())
		return
	}

	for _, c := range game.info.Contracts {
		msg, err := trx.AsMessage(types.NewLondonSigner(big.NewInt(int64(sp.chainID))), big.NewInt(0))
		if err != nil {
			log.Error("[%s:%v]AsMessage error:%s", trx.Hash().Hex(), trx.Type(), err.Error())
			return nil // success when can not AsMessage
		}
		if trx.To() == nil {
			return nil // done with 0x0000...000
		}
		//log.Info("[%s]:c[%s]  to[%s]", trx.Hash().Hex(), c.Address, trx.To().Hex())
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
			_ = method
			// trxModel := &db.Transaction{
			// 	GameID:    game.info.ID,
			// 	Timestamp: blk.Header().Time,
			// 	ID:        trx.Hash().Hex(),
			// 	From:      msg.From().Hex(),
			// 	To:        trx.To().Hex(),
			// 	BlockNum:  blk.NumberU64(),
			// 	Method:    method,
			// }
			// log.Info("trxModel:%v", trxModel)

			// ctx, cancel := context.WithTimeout(sp.ctx, 3*time.Second)
			// defer cancel()

			// opt := mngopts.FindOneAndReplace()
			// opt.SetUpsert(true)

			// filter := bson.M{
			// 	"_id": trxModel.ID,
			// }

			// rst := gameTbl.FindOneAndReplace(ctx, filter, trxModel, opt)
			// if rst.Err() != nil {
			// 	if !strings.Contains(rst.Err().Error(), "no documents in result") {
			// 		log.Error("update transaction[%s] error: %s", trxModel.ID, rst.Err().Error())
			// 		return err // return to show error
			// 	}
			// }
		}
	}
	return nil
}

func (sp *ETHSpider) dealBlock(number uint64) (err error) {
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

func (sp *ETHSpider) goBackward() {
	log.Info("go backward")
	sp.tailBlock = sp.topBlock
	interval := time.Duration(sp.backwardInterval * float32(time.Second))

	for {
		err := sp.dealBlock(sp.tailBlock)
		if err != nil {
			log.Error("deal block[%d]:%s", sp.headBlock, err.Error())
			time.Sleep(interval)
			continue
		}
		if sp.tailBlock%10 == 0 {
			log.Info("backfoward to:%v", sp.tailBlock)
		}
		sp.tailBlock -= 1

		if sp.tailBlock <= sp.bottomBlock {
			break
		}
		time.Sleep(interval)
	}

	log.Info("done all backfoward")
}
