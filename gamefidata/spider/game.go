package spider

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

type Game struct {
	ctx           context.Context
	cancelFun     context.CancelFunc
	info          GameInfo
	ethcli        *ethclient.Client
	dbClient      *mongo.Client
	db            *mongo.Database
	trxTbl        *mongo.Collection
	minBlock      uint64
	blockHeight   uint64
	curLowBlock   uint64
	curHeighBlock uint64
}

func NewGame(info GameInfo) *Game {
	gm := &Game{
		info: info,
	}
	return gm
}

// func (gm *Game) Run() (err error) {
// 	var number uint64
// 	for _, c := range gm.info.Contracts {
// 		if c.StartBlock < number ||
// 			(uint64(0) == number) {
// 			number = c.StartBlock
// 		}
// 	}
// 	gm.minBlock = number
// 	log.Info("Deal for game:%s", gm.info.Name)
// 	go gm.updateBlockHeightLoop()

// 	gm.loadCurBlocks()

// 	for {
// 		if gm.curHeighBlock < gm.blockHeight {
// 			err := gm.dealBlock(gm.curHeighBlock)
// 			if err == nil {
// 				gm.curHeighBlock += 1
// 			}
// 		}
// 		if gm.curLowBlock >= gm.minBlock {
// 			err := gm.dealBlock(gm.curLowBlock)
// 			if err == nil {
// 				gm.curLowBlock -= 1
// 			}
// 		}
// 	}
// }

// func (gm *Game) dealBlock(number uint64) (err error) {
// 	blk, err := gm.ethcli.BlockByNumber(gm.ctx, big.NewInt(int64(number)))
// 	if err != nil {
// 		log.Error("get block[%d] error:%s", number, err.Error())
// 		//time.Sleep(5 * time.Second)
// 		if strings.Contains(err.Error(), "block header indicates no transactions") {
// 			return nil
// 		}
// 		return
// 	}
// 	for _, trx := range blk.Transactions() {
// 		for _, c := range gm.info.Contracts {
// 			msg, err := trx.AsMessage(types.NewEIP155Signer(big.NewInt(int64(gm.info.ChainID))), big.NewInt(0))
// 			if err != nil {
// 				return nil // success when can not AsMessage
// 			}
// 			if trx.To() == nil {
// 				return nil // done with 0x0000...000
// 			}
// 			if c.Address == trx.To().Hex() {
// 				log.Info("[%s] %s send transaction to contract:%v %v:%v", trx.Hash().Hex(), msg.From().Hex(), trx.To().Hex(), blk.Header().Time, blk.NumberU64())
// 				inputData := fmt.Sprintf("0x%x", trx.Data())
// 				method := ""
// 				if len(inputData) > 10 {
// 					methodTxt := inputData[:10]

// 					method, err = fourbyte.DB.Get(methodTxt)
// 					if err != nil {
// 						// TODO: retry
// 						log.Error("Get fourbyte error:%s", err.Error())
// 					}
// 				}
// 				trxModel := &db.Transaction{
// 					GameID:    gm.info.ID,
// 					Timestamp: blk.Header().Time,
// 					ID:        trx.Hash().Hex(),
// 					From:      msg.From().Hex(),
// 					To:        trx.To().Hex(),
// 					BlockNum:  blk.NumberU64(),
// 					Method:    method,
// 				}
// 				log.Info("trxModel:%v", trxModel)

// 				ctx, cancel := context.WithTimeout(gm.ctx, 3*time.Second)
// 				defer cancel()
// 				rst, err := gm.trxTbl.InsertOne(ctx, trxModel)
// 				if err != nil {
// 					if !strings.Contains(err.Error(), "duplicate key error collection") {
// 						log.Error("insert transaction[%s] error: %s", trxModel.ID, err.Error())
// 						return nil // success when duplicate
// 					}
// 				} else {
// 					log.Info("[SUCC]insert transaction[%s] ", rst.InsertedID)
// 				}

// 			}
// 		}
// 	}
// 	return nil
// }

// func (gm *Game) updateBlockHeightLoop() {

// 	ticker := time.NewTicker(5000 * time.Millisecond) // for 5sec
// 	done := make(chan bool)

// 	for {
// 		select {
// 		case <-done:
// 			return
// 		case <-ticker.C:
// 			gm.updateBlockHeight()
// 		}
// 	}

// }

// func (gm *Game) updateBlockHeight() (err error) {
// 	height, err := gm.ethcli.BlockNumber(gm.ctx)
// 	if err != nil {
// 		log.Error("get block height error:%s", err)
// 		return
// 	}
// 	atomic.StoreUint64(&gm.blockHeight, height)
// 	return
// }

// func (gm *Game) loadCurBlocks() (err error) {
// 	pipeline := []bson.M{{
// 		"$group": bson.M{
// 			"_id": "$to",
// 			"maxHeight": bson.M{
// 				"$max": "$blocknum",
// 			},
// 		},
// 	}}

// 	cursor, err := gm.trxTbl.Aggregate(gm.ctx, pipeline)
// 	if err != nil {
// 		log.Error("Get max record error:%s", err.Error())
// 		return
// 	}

// 	for cursor.Next(gm.ctx) {
// 		if cursor.Current == nil {
// 			gm.curHeighBlock = gm.blockHeight
// 			log.Info("set  curHeighBlock to gm.blockHeight:%d", gm.blockHeight)
// 			break
// 		}
// 		v := cursor.Current.Lookup("maxHeight")
// 		if gm.curHeighBlock >= uint64(v.AsInt64()) ||
// 			gm.curHeighBlock == 0 {
// 			gm.curHeighBlock = uint64(v.AsInt64())
// 		}
// 	}
// 	if gm.curHeighBlock == 0 {
// 		gm.curHeighBlock = gm.blockHeight
// 	}

// 	pipeline = []bson.M{{
// 		"$group": bson.M{
// 			"_id": "$to",
// 			"maxHeight": bson.M{
// 				"$min": "$blocknum",
// 			},
// 		},
// 	}}

// 	cursor, err = gm.trxTbl.Aggregate(gm.ctx, pipeline)
// 	if err != nil {
// 		log.Error("Get min record error:%s", err.Error())
// 		return
// 	}
// 	for cursor.Next(gm.ctx) {
// 		if cursor.Current == nil {
// 			gm.curHeighBlock = gm.blockHeight
// 			log.Info("set  curHeighBlock to gm.blockHeight:%d", gm.blockHeight)
// 			break
// 		}
// 		v := cursor.Current.Lookup("maxHeight")
// 		if gm.curLowBlock <= uint64(v.AsInt64()) {
// 			gm.curLowBlock = uint64(v.AsInt64())
// 		}
// 	}
// 	if gm.curLowBlock == 0 {
// 		gm.curLowBlock = gm.blockHeight
// 	}

// 	log.Info("curHeighBlock:%v curLowBlock:%v", gm.curHeighBlock, gm.curLowBlock)
// 	return
// }
