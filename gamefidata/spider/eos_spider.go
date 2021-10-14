package spider

import (
	"context"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/eoscanada/eos-go"
	"github.com/gametaverse/gamefidata/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type EOSSpider struct {
	ctx              context.Context
	cancelFun        context.CancelFunc
	forwardInterval  float32
	backwardInterval float32
	chainID          int
	chain            string
	eoscli           *eos.API
	dbClient         *mongo.Client
	db               *mongo.Database
	games            []*Game
	topBlock         uint32
	headBlock        uint32
	tailBlock        uint32
	bottomBlock      uint32
	mongoURI         string
	rpcAddr          string
	backward         bool
	monitorField     string
	monitorTbl       *mongo.Collection
}

func (sp *EOSSpider) Init() (err error) {
	log.Info("init")
	sp.ctx, sp.cancelFun = context.WithCancel(context.Background())

	sp.eoscli = eos.New(sp.rpcAddr)

	sp.monitorField = db.MonitorFieldName + "_" + sp.chain
	err = sp.initDB(sp.mongoURI)
	if err != nil {
		log.Error("Init mongon error:%s", err.Error())
		return err
	}

	return
}

func (sp *EOSSpider) initDB(URI string) (err error) {
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

func (sp *EOSSpider) Run() {
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

func (sp *EOSSpider) loadTopBlock() (err error) {
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
		sp.topBlock = uint32(m.TopBlock)
		break
	}

	if sp.topBlock == 0 {
		// No such record
		log.Info("go to get block")
		sp.topBlock, err = sp.getBlockHeight()
		if err != nil {
			log.Error("get block height error:", err.Error())
			return err
		}
	}
	log.Info("topBlock:%v", sp.topBlock)
	return
}

func (sp *EOSSpider) getBlockHeight() (height uint32, err error) {
	infoRsp, err := sp.eoscli.GetInfo(sp.ctx)
	if err != nil {
		log.Error("get block height error:%s", err)
		return
	}
	log.Info("block height:%v", infoRsp.HeadBlockNum)
	height = infoRsp.HeadBlockNum
	return
}

func (sp *EOSSpider) goForward() {
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

func (sp *EOSSpider) storeTopBlock(number uint32) (err error) {
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

func (sp *EOSSpider) dealGame(game *Game, blk *eos.BlockResp, trx *eos.TransactionReceipt) (err error) {

	gameTbl := sp.db.Collection("t_" + game.info.ID)
	if gameTbl == nil {
		log.Error("create collection for game[%s] error:%s", game.info.ID, err.Error())
		return
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
		for i, act := range strx.Actions {
			cName := act.Account
			if c.Address != cName.String() {
				continue
			}
			method := act.Name
			from := ""
			for j, auth := range act.Authorization {
				if auth.Permission.String() == "active" {
					from = auth.Actor.String()

					log.Info("trx[%v_%d_%d]:context:%v:method:%v:from:%v", trx.Transaction.ID, i, j, cName, method, from)

					trxModel := &db.Transaction{
						GameID:    game.info.ID,
						Timestamp: uint64(blk.Timestamp.Unix()),
						ID:        trx.Transaction.ID.String(),
						From:      from,
						To:        cName.String(),
						BlockNum:  uint64(blk.BlockNum),
						Method:    method.String(),
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

		}

	}
	return nil
}

func (sp *EOSSpider) dealBlock(number uint32) (err error) {
	blk, err := sp.eoscli.GetBlockByNum(sp.ctx, number)
	if err != nil {
		//log.Error("get block[%d] error:%s", number, err.Error())
		if strings.Contains(err.Error(), "block header indicates no transactions") {
			return nil
		}
		return
	}
	for _, trx := range blk.Transactions {
		for _, g := range sp.games {
			err = sp.dealGame(g, blk, &trx)
			if err != nil {
				log.Error("deal game error: %s", err.Error())
				return err // for show errors
			}
		}
	}
	return nil
}

func (sp *EOSSpider) goBackward() {
	sp.tailBlock = sp.topBlock
	interval := time.Duration(sp.backwardInterval * float32(time.Second))
	timer := time.NewTimer(interval)
	for range timer.C {
		err := sp.dealBlock(sp.tailBlock)
		if err != nil {
			log.Error("deal block[%d] error: %s", sp.tailBlock, err.Error())
		} else {
			sp.tailBlock -= 1
		}
		if sp.tailBlock%100 == 0 {
			log.Info("backfoward to :%v", sp.tailBlock)
		}
		if sp.tailBlock <= sp.bottomBlock {
			break
		}
		timer.Reset(interval)
	}
	log.Info("done all backfoward")
}
