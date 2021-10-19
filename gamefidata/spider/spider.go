package spider

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Spider struct {
	ctx       context.Context
	antenna   Antennaer
	cancelFun context.CancelFunc
	games     []*Game
	opts      options

	dbClient  *mongo.Client
	db        *mongo.Database
	topBlock  uint64
	headBlock uint64
	tailBlock uint64

	backward     bool
	monitorField string
	monitorTbl   *mongo.Collection
	dauTbl       *mongo.Collection
	countTbl     *mongo.Collection
}

func NewSpider(games []*Game, opts options, backward bool) *Spider {
	s := &Spider{
		games:    games,
		opts:     opts,
		backward: backward,
	}
	switch opts.Chain {
	case "polygon", "eth", "bsc":
		s.antenna = NewETHAntenna()
	case "wax":
		s.antenna = NewEOSAntenna()
	}
	return s
}

func (sp *Spider) Init() (err error) {
	log.Info("BaseSpider init")
	sp.ctx, sp.cancelFun = context.WithCancel(context.Background())

	err = sp.antenna.Init(sp.opts.RPCAddr, sp.opts.ChainID)
	if err != nil {
		log.Error("Antenna init error:%s", err.Error())
		return
	}

	sp.monitorField = db.MonitorFieldName + "_" + sp.opts.Chain
	err = sp.initDB(sp.opts.MongoURI)
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

func (sp *Spider) Run(starting chan struct{}) {
	err := sp.loadTopBlock()
	if err != nil {
		log.Error("load top block:", err.Error())
		return
	}

	if sp.backward {
		sp.goBackward(starting)
	} else {
		sp.goForward(starting)
	}
}

func (sp *Spider) loadTopBlock() (err error) {
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

func (sp *Spider) getBlockHeight() (height uint64, err error) {
	height, err = sp.antenna.GetBlockHeight(sp.ctx)
	if err != nil {
		log.Error("get block height error:%s", err)
		return
	}
	log.Info("block height:%v", height)
	return
}

func (sp *Spider) getTrxFromBlocks(start uint64, count int) (trxes []*Transaction, err error) {
	var wg sync.WaitGroup
	type result struct {
		trxes []*Transaction
		err   error
	}
	rstChan := make(chan *result, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i uint64) {
			trx, err := sp.antenna.GetTrxByNum(sp.ctx, i)
			if err != nil {
				log.Error("get block error:%s", err.Error())
			}
			rst := &result{
				trxes: trx,
				err:   err,
			}
			rstChan <- rst
			wg.Done()
		}(start + uint64(i))
	}
	wg.Wait()
	i := 0
	for rst := range rstChan {
		if rst.err != nil {
			return nil, ErrGetBlock
		}
		trxes = append(trxes, rst.trxes...)
		i++
		if i == count {
			break
		}
	}
	return trxes, nil
}

func (sp *Spider) goForward(starting chan struct{}) {
	<-starting
	sp.headBlock = sp.topBlock
	log.Info("go forward from %v", sp.headBlock)
	count := sp.opts.ForwardWorks
	for {
		s1 := time.Now()
		trxes, err := sp.getTrxFromBlocks(sp.headBlock, count)
		if err != nil {
			log.Error("get %d block[%d]:%s", count, sp.headBlock, err.Error())
			time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
			continue
		}
		err = sp.dealTrxes(trxes)
		if err != nil {
			log.Error("deal %d block[%d]:%s", count, sp.headBlock, err.Error())
			time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
			continue
		}
		if (sp.headBlock/uint64(count))%10 == 0 {
			sp.storeTopBlock(sp.headBlock)
		}
		sp.headBlock += uint64(count)
		s2 := time.Now()
		d := s2.Sub(s1)
		log.Info("%d blocks cost d is %v", count, d)
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
	_, err = sp.monitorTbl.UpdateByID(ctx, sp.monitorField, update, opt)
	if err != nil {
		log.Error("Update top block error: ", err.Error())
		return
	}
	log.Info("Update top block to:%d ", number)
	return
}

func (sp *Spider) dealTrxes(trxes []*Transaction) (err error) {
	for _, g := range sp.games {
		err = sp.dealTrxes4Game(g, trxes)
		if err != nil {
			log.Error("dealBlocks4Game error:%s", err.Error())
		}
	}
	return nil
}

func (sp *Spider) dealTrxes4Game(game *Game, trxes []*Transaction) (err error) {
	var actions = make(map[string]*db.Action)
	for _, trx := range trxes {
		as, err := sp.antenna.DealTrx4Game(game, trx)
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

func (sp *Spider) insertDAU(actions map[string]*db.Action) (err error) {
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

func (sp *Spider) clearTick(timestamp uint64) (err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 3*time.Second)
	defer cancel()

	opts := mngopts.Update()
	opts.SetUpsert(false)
	filter := bson.M{
		"ts": timestamp,
	}

	update := bson.M{
		"$set": bson.M{
			"count": 0,
		},
	}
	rst, err := sp.countTbl.UpdateMany(ctx, filter, update, opts)
	if err != nil {
		log.Error("UpdateMany  error: %s", err.Error())
		return err // return to show error
	}
	_ = rst
	return
}

func (sp *Spider) insertCount(actions map[string]*db.Action) (err error) {
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

func (sp *Spider) goBackward(starting chan struct{}) {
	sp.tailBlock = sp.topBlock
	log.Info("go backforward from :%v", sp.tailBlock)
	count := sp.opts.BackwardWorks
	log.Info("opts:%v", sp.opts)
	interval := time.Duration(sp.opts.BackwardInterval * float32(time.Second))
	log.Info("interval:%v", interval)
	var minTimeStamp uint64 = math.MaxUint64
	for {
		trxes, err := sp.getTrxFromBlocks(sp.tailBlock, count)
		if err != nil {
			log.Error("get %d block[%d]:%s", count, sp.tailBlock, err.Error())
			time.Sleep(time.Duration(sp.opts.BackwardInterval * float32(time.Second)))
			continue
		}
		oldMinTS := minTimeStamp
		for _, trx := range trxes {
			if trx.timestamp < minTimeStamp {
				minTimeStamp = trx.timestamp
			}
		}
		// the interval won't bigger than 24h
		if oldMinTS > minTimeStamp {
			err = sp.clearTick(minTimeStamp)
			if err != nil {
				log.Info("clear tick error:%s", err.Error())
				return
			}
			if oldMinTS == math.MaxUint64 {
				starting <- struct{}{}
			}
		}
		err = sp.dealTrxes(trxes)
		if err != nil {
			log.Error("deal %d block[%d]:%s", count, sp.tailBlock, err.Error())
			time.Sleep(time.Duration(sp.opts.BackwardInterval * float32(time.Second)))
			continue
		}
		if (sp.tailBlock/uint64(count))%10 == 0 {
			log.Info("back to block to: %v", sp.tailBlock)
		}
		sp.tailBlock -= uint64(count)
		if sp.tailBlock <= sp.opts.BottomBlock {
			break
		}
		time.Sleep(interval)
	}
	log.Info("done all backfoward")
}
