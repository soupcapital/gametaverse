package spider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/eoscanada/eos-go"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/utils"
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
	forwardWorks     int
	backwardWorks    int
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
	dauTbl           *mongo.Collection
	countTbl         *mongo.Collection
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

func (sp *EOSSpider) getBlocks(start uint32, count uint32) (blocks []*eos.BlockResp, err error) {
	var wg sync.WaitGroup
	rstChan := make(chan *eos.BlockResp, count)
	for i := uint32(0); i < count; i++ {
		wg.Add(1)
		go func(i uint32) {
			blk, err := sp.eoscli.GetBlockByNum(sp.ctx, i)
			if err != nil {
				log.Error("get block error:%s", err.Error())
			}
			rstChan <- blk
			wg.Done()
		}(start + i)
	}
	wg.Wait()
	for blk := range rstChan {
		if blk == nil {
			return blocks, ErrGetBlock
		}
		blocks = append(blocks, blk)
		if len(blocks) == int(count) {
			break
		}
	}
	return blocks, nil
}

func (sp *EOSSpider) goForward() {
	log.Info("go forward")
	sp.headBlock = sp.topBlock
	count := uint32(sp.forwardWorks)
	i := uint32(0)
	for {
		s1 := time.Now()
		blocks, err := sp.getBlocks(sp.headBlock, count)
		if err != nil {
			log.Error("get %d block[%d]:%s", count, sp.headBlock, err.Error())
			time.Sleep(time.Duration(sp.forwardInterval * float32(time.Second)))
			continue
		}
		err = sp.dealBlocks(blocks)
		if err != nil {
			log.Error("deal %d block[%d]:%s", count, sp.headBlock, err.Error())
			time.Sleep(time.Duration(sp.forwardInterval * float32(time.Second)))
			continue
		}
		if i%(10/count) == 0 {
			sp.storeTopBlock(sp.headBlock)
		}
		sp.headBlock += count
		i++
		s2 := time.Now()
		d := s2.Sub(s1)
		log.Info("%d blocks cost d is %v", count, d)
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
	log.Info("Update top block to:%d ", number)
	return
}

func (sp *EOSSpider) dealTrx4Game(game *Game, blk *eos.BlockResp, trx *eos.TransactionReceipt) (actions []*db.Action, err error) {
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
						Timestamp: utils.StartSecForDay(uint64(blk.Timestamp.Unix())),
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

func (sp *EOSSpider) dealBlocks(blocks []*eos.BlockResp) (err error) {
	for _, g := range sp.games {
		err = sp.dealBlocks4Game(g, blocks)
		if err != nil {
			log.Error("dealBlocks4Game error:%s", err.Error())
		}
	}
	return nil
}

func (sp *EOSSpider) dealBlocks4Game(game *Game, blocks []*eos.BlockResp) (err error) {
	var actions = make(map[string]*db.Action)
	for _, blk := range blocks {
		for _, trx := range blk.Transactions {
			as, err := sp.dealTrx4Game(game, blk, &trx)
			if err != nil {
				log.Error("deal game error: %s", err.Error())
				return err // for show errors
			}
			for _, a := range as {
				key := fmt.Sprintf("%v_%v_%v", a.GameID, a.Timestamp, a.User)
				if _, ok := actions[key]; !ok {
					actions[key] = a
				} else {
					actions[key].Count += a.Count
				}
			}
		}
	}

	if len(actions) == 0 {
		return
	}

	err = sp.insertDAU(actions)
	if err != nil {
		log.Error("insert DAU error:%s", err.Error())
		return err
	}
	err = sp.insertCount(actions)
	if err != nil {
		log.Error("insert DAU error:%s", err.Error())
		return err
	}
	return
}

func (sp *EOSSpider) insertDAU(actions map[string]*db.Action) (err error) {
	var docs []interface{}
	for _, a := range actions {
		doc := db.DAU{
			GameID:    a.GameID,
			Timestamp: a.Timestamp,
			User:      a.User,
		}
		docs = append(docs, doc)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	opts := mngopts.InsertMany()
	opts.SetOrdered(false)

	rst, err := sp.dauTbl.InsertMany(ctx, docs, opts)
	if err != nil {
		log.Error("InsertMany  error: %s", err.Error())
		return err // return to show error
	}
	_ = rst
	return
}

func (sp *EOSSpider) insertCount(actions map[string]*db.Action) (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 3*time.Second)
	defer cancel()

	counts := make(map[string]*db.Action)

	for _, a := range actions {
		key := fmt.Sprintf("%v_%v", a.GameID, a.Timestamp)
		if _, ok := counts[key]; !ok {
			counts[key] = a
		} else {
			counts[key].Count += a.Count
		}
	}
	for _, a := range counts {
		opts := mngopts.Update()
		opts.SetUpsert(true)
		filter := bson.M{
			"game": a.GameID,
			"ts":   a.Timestamp,
		}

		update := bson.M{
			"$inc": bson.M{
				"count": a.Count,
			},
		}
		rst, err := sp.countTbl.UpdateMany(ctx, filter, update, opts)
		if err != nil {
			log.Error("UpdateMany  error: %s", err.Error())
			return err // return to show error
		}
		_ = rst
	}

	return nil
}

func (sp *EOSSpider) goBackward() {
	sp.tailBlock = sp.topBlock
	log.Info("go backforward from :%v", sp.tailBlock)
	count := uint32(sp.backwardWorks)
	interval := time.Duration(sp.backwardInterval * float32(time.Second))
	log.Info("interval:%v", interval)
	i := uint32(0)
	for {
		blocks, err := sp.getBlocks(sp.tailBlock-count, count)
		if err != nil {
			log.Error("get %d block[%d]:%s", count, sp.headBlock, err.Error())
			time.Sleep(interval)
			continue
		}
		err = sp.dealBlocks(blocks)
		if err != nil {
			log.Error("deal %d block[%d]:%s", count, sp.tailBlock, err.Error())
			time.Sleep(interval)
			continue
		}
		if i%(10/count) == 0 {
			log.Info("backfoward to:%v", sp.tailBlock)
		}
		sp.tailBlock -= count
		i++
		if sp.tailBlock <= sp.bottomBlock {
			break
		}
		time.Sleep(interval)
	}

	log.Info("done all backfoward")
}
