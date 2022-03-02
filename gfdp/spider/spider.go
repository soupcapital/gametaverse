package spider

import (
	"context"
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

	curBlock uint64

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
	// ctx := sp.ctx
	// if err := sp.dbConn.Exec(ctx, `DROP TABLE IF EXISTS t_txs`); err != nil {
	// 	log.Error("drop table error:%s", err.Error())
	// 	return err
	// }
	// err = sp.dbConn.Exec(ctx, `
	// 	CREATE TABLE IF NOT EXISTS t_txs (
	// 		tx_hash String NOT NULL,
	// 		ts DateTime,
	// 		from String,
	// 		to String,
	// 		data String
	// 	)   ENGINE = ReplacingMergeTree()
	// 		ORDER BY  (ts, to, tx_hash)
	// 		PRIMARY KEY (ts,to);
	// `)
	// if err != nil {
	// 	log.Error("creat table error:%s", err.Error())
	// 	return err
	// }
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

	sp.curBlock, err = sp.loadMaxBlock()
	if err != nil {
		log.Error("get max block error:%s", err.Error())
		return
	}
	if sp.curBlock == 0 {
		sp.curBlock = sp.opts.BottomBlock
	}
	log.Info("curBlock:%d", sp.curBlock)
	sp.run(ctx)
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
	//log.Info("Insert into [%d] tx", len(allTx))
	batch, err := sp.dbConn.PrepareBatch(sp.ctx, "INSERT INTO t_txs")
	if err != nil {
		log.Info("PrepareBatch error:%s", err.Error())
		return err
	}
	for _, tx := range allTx {
		if err = batch.Append(tx.Hash,
			tx.Timestamp,
			tx.Block,
			tx.From,
			tx.To,
			tx.Data); err != nil {
			log.Error("batch append error:%s", err.Error())
			continue
		}
	}
	err = batch.Send()
	return
}

func (sp *Spider) run(ctx context.Context) {
	count := sp.opts.ForwardWorks
	for {
		select {
		case <-ctx.Done():
			log.Info("forward got stop guard")
			return
		default:
			trxes, err := sp.getTrxFromBlocks(sp.curBlock, count)
			if err != nil {
				log.Error("get %d block[%d]:%s", count, sp.curBlock, err.Error())
				time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
				continue
			}
			//log.Info("got %d txes from block[%d]", len(trxes), sp.curBlock)
			err = sp.dealTrxes(trxes)
			if err != nil {
				log.Error("deal %d block[%d]:%s", count, sp.curBlock, err.Error())
				time.Sleep(time.Duration(sp.opts.ForwardInterval * float32(time.Second)))
				continue
			}
			sp.curBlock += uint64(count)
			if (sp.curBlock/uint64(count))%10 == 0 {
				log.Info("back to block to: %v", sp.curBlock)
			}
		}
	}
}
