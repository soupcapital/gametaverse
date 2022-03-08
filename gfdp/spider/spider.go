package spider

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/db"
)

type Spider struct {
	ctx       context.Context
	antenna   Antennaer
	cancelFun context.CancelFunc
	opts      options
	wg        sync.WaitGroup

	forwardBlock  uint64
	backwardBlock uint64

	stopGuard chan struct{}
	dbConn    clickhouse.Conn
}

func NewSpider(opts options) *Spider {
	s := &Spider{
		opts:      opts,
		stopGuard: make(chan struct{}),
	}
	switch opts.Chain {
	case "polygon", "eth", "bsc":
		s.antenna = NewETHAntenna()
		// case "avax":
		// 	s.antenna = NewAvaxAntenna()
		// case "wax":
		// 	s.antenna = NewEOSAntenna()
		// case "solana":
		// 	s.antenna = NewSolanaAntenna()
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

	if err = sp.initDb(); err != nil {
		log.Error("init db error:%s", err.Error())
		return
	}
	log.Info("init db success")

	return
}

func (sp *Spider) initDb() (err error) {
	sp.dbConn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "d_gamefidata",
			Username: "default",
			Password: "",
		},
		//Debug:           true,
		DialTimeout:     time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})

	return
}

func (sp *Spider) Stop() {
	log.Info("before channle")
	sp.stopGuard <- struct{}{}
	log.Info("after channle")
}

func (sp *Spider) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	var err error

	maxBlock, err := sp.loadMaxBlock()
	if err != nil {
		log.Error("get max block error:%s", err.Error())
		return
	}
	minBlock, err := sp.loadMinBlock()
	if err != nil {
		log.Error("get max block error:%s", err.Error())
		return
	}

	ctxMin, cancel := context.WithTimeout(sp.ctx, 60*time.Second)
	defer cancel()
	curHeight, err := sp.antenna.GetBlockHeight(ctxMin)
	if err != nil {
		log.Error("get cur block error:%s", err.Error())
		return
	}

	if minBlock == 0 {
		if maxBlock == 0 {
			minBlock = curHeight
			maxBlock = curHeight
		} else {
			log.Error("minBlock and maxBlock are not all0")
			return
		}
	}
	if minBlock == maxBlock {
		minBlock -= 1
	}
	sp.backwardBlock = minBlock
	sp.forwardBlock = maxBlock
	log.Info("back from[%d] and forward from[%d]", sp.backwardBlock, sp.forwardBlock)
	sp.routine(ctx, sp.goForward)
	sp.routine(ctx, sp.goBackward)

	sp.wg.Wait()
}

func (sp *Spider) routine(ctx context.Context, fn func(context.Context, *sync.WaitGroup)) {
	sp.wg.Add(1)
	go fn(ctx, &sp.wg)
}

func (sp *Spider) loadMaxBlock() (block uint64, err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()
	if err = sp.dbConn.QueryRow(ctx, "SELECT MAX(blk_num) FROM t_txs").Scan(&block); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}
	return
}

func (sp *Spider) loadMinBlock() (block uint64, err error) {
	ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Second)
	defer cancel()
	if err = sp.dbConn.QueryRow(ctx, "SELECT MIN(blk_num) FROM t_txs").Scan(&block); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}
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
			ctx, cancel := context.WithTimeout(sp.ctx, 5*time.Minute)
			defer cancel()
			trx, err := sp.antenna.GetTrxByNum(ctx, i)
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
			return nil, rst.err
		}
		trxes = append(trxes, rst.trxes...)
		i++
		if i == count {
			break
		}
	}
	return trxes, nil
}

func (sp *Spider) dealTrxes(trxes []*Transaction) (err error) {
	var allTx []*db.Transaction
	for _, tx := range trxes {
		txes, err := sp.antenna.DealTrx(tx)
		if err != nil {
			log.Info("deal tx error:%s", err.Error())
			continue
		}
		allTx = append(allTx, txes...)
	}
	if len(allTx) == 0 {
		return nil
	}
	batch, err := sp.dbConn.PrepareBatch(sp.ctx, "INSERT INTO t_txs")
	if err != nil {
		log.Info("PrepareBatch error:%s", err.Error())
		return err
	}
	for _, tx := range allTx {
		if err = batch.Append(
			tx.Timestamp,
			tx.Block,
			tx.From,
			tx.To); err != nil {
			log.Error("batch append error:%s", err.Error())
			continue
		}
	}
	err = batch.Send()
	return
}

func (sp *Spider) goForward(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	count := sp.opts.ForwardWorks
	for {
		select {
		case <-ctx.Done():
			log.Info("forward got stop guard")
			return
		default:
			trxes, err := sp.getTrxFromBlocks(sp.forwardBlock, count)
			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					log.Error("get %d block[%d]:%s", count, sp.forwardBlock, err.Error())
				}
				time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
				continue
			}
			err = sp.dealTrxes(trxes)
			if err != nil {
				log.Error("deal %d block[%d]:%s", count, sp.forwardBlock, err.Error())
				time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
				continue
			}
			sp.forwardBlock += uint64(count)
			if (sp.forwardBlock/uint64(count))%10 == 0 {
				log.Info("[PROC] forward to block to: %v", sp.forwardBlock)
			}
		}
	}
}

func (sp *Spider) goBackward(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	count := sp.opts.BackwardWorks
	for {
		select {
		case <-ctx.Done():
			log.Info("backward got stop guard")
			return
		default:
			trxes, err := sp.getTrxFromBlocks(sp.backwardBlock-uint64(count), count)
			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					log.Error("get %d block[%d]:%s", count, sp.backwardBlock, err.Error())
					time.Sleep(time.Duration(sp.opts.BackwardInterval * float32(time.Second)))
					continue
				}
			}
			err = sp.dealTrxes(trxes)
			if err != nil {
				log.Error("deal %d block[%d]:%s", count, sp.backwardBlock, err.Error())
				time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
				continue
			}
			sp.backwardBlock -= uint64(count)
			if (sp.backwardBlock/uint64(count))%10 == 0 {
				log.Info("[PROC]backward to block to: %v", sp.backwardBlock)
			}

			if sp.backwardBlock <= sp.opts.BottomBlock {
				log.Info("Backward to bottom")
				return
			}
		}
	}
}
