package rpc

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/rpc/pb"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedDBProxyServer
	ctx       context.Context
	cancelFun context.CancelFunc
	opts      options
	listener  net.Listener
	rpcSvr    *grpc.Server
	dbConn    driver.Conn
	dbCache   *cache
}

func NewServer() (svr *Server) {
	svr = &Server{
		dbCache: NewCache(),
	}
	return svr
}

func (svr *Server) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&svr.opts)
	}

	svr.ctx, svr.cancelFun = context.WithCancel(context.Background())
	log.Info("opts:%v", svr.opts)
	if svr.dbConn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{svr.opts.DbAddr},
		Auth: clickhouse.Auth{
			Database: svr.opts.DbName,
			Username: svr.opts.DbUser,
			Password: svr.opts.DbPasswd,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time":        60,
			"max_query_size":            10000000000000,
			"max_ast_elements":          5000000000,
			"max_expanded_ast_elements": 5000000000,
		},
		//Debug: true,
	}); err != nil {
		log.Error("open db error:%s", err.Error())
		return
	}

	svr.rpcSvr = grpc.NewServer()
	pb.RegisterDBProxyServer(svr.rpcSvr, svr)
	return
}

func (svr *Server) Run() (err error) {
	svr.listener, err = net.Listen("tcp", svr.opts.ListenAddr)
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	err = svr.rpcSvr.Serve(svr.listener)
	if err != nil {
		log.Error("Serve Error:%s", err.Error())
	}
	return
}

func (svr *Server) loadMaxTimestamp(table string) (ts time.Time, err error) {
	ctx, cancel := context.WithTimeout(svr.ctx, 5*time.Second)
	defer cancel()
	var block uint64
	sql := fmt.Sprintf("SELECT MAX(blk_num) FROM %s", table)
	if err = svr.dbConn.QueryRow(ctx, sql).Scan(&block); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}
	sql = fmt.Sprintf("SELECT MAX(ts) FROM %s WHERE blk_num = %d", table, block)
	if err = svr.dbConn.QueryRow(ctx, sql).Scan(&ts); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}
	return
}

func (svr *Server) loadMinTimestamp(table string) (ts time.Time, err error) {
	ctx, cancel := context.WithTimeout(svr.ctx, 5*time.Second)
	defer cancel()
	var block uint64
	sql := fmt.Sprintf("SELECT MIN(blk_num) FROM %s", table)
	if err = svr.dbConn.QueryRow(ctx, sql).Scan(&block); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}
	sql = fmt.Sprintf("SELECT MIN(ts) FROM %s WHERE blk_num = %d ", table, block)
	if err = svr.dbConn.QueryRow(ctx, sql).Scan(&ts); err != nil {
		log.Error("query error:%s", err.Error())
		return
	}

	return
}

func (svr *Server) getBlockScope(chains []pb.Chain) (min, max time.Time, err error) {
	log.Info("chains:%#v", chains)
	for _, chain := range chains {
		table, err := getTableName(chain)
		if err != nil {
			return min, max, err
		}
		tmin, err := svr.loadMinTimestamp(table)
		if err != nil {
			return min, max, err
		}
		tmax, err := svr.loadMaxTimestamp(table)
		if err != nil {
			return min, max, err
		}

		if max.IsZero() {
			max = tmax
		} else {
			if tmax.Unix() < max.Unix() {
				max = tmax
			}
		}

		if min.IsZero() {
			min = tmin
		} else {
			if tmin.Unix() > min.Unix() {
				min = tmin
			}
		}
	}
	err = nil
	return
}

func (svr *Server) Dau(ctx context.Context, req *pb.GameReq) (rsp *pb.DauRsp, err error) {
	var count uint64
	key := fmt.Sprintf("%v-%v", req.Start, req.End)
	for _, c := range req.Contracts {
		key += fmt.Sprintf("%v:%v", c.Chain, c.Address)
	}
	keyMd5 := md5.Sum([]byte(key))
	key = hex.EncodeToString(keyMd5[:])

	if count, err = svr.dbCache.getDau(key); err == nil {
		rsp = &pb.DauRsp{
			Dau: count,
		}
		err = nil
		return
	}

	var contracts map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	var chains []pb.Chain
	for _, c := range req.Contracts {
		contracts[c.Chain] = append(contracts[c.Chain], c.Address)
	}
	for k := range contracts {
		chains = append(chains, k)
	}

	sql := ""

	if len(contracts) == 1 {
		// only one chain
		for k, v := range contracts {
			tblName, _ := getTableName(k)
			sql = fmt.Sprintf("SELECT countDistinct(from) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			toSql := " AND ("
			s := ""
			for _, c := range v {
				toEqul := fmt.Sprintf(" %s (to = '%s')", s, c)
				s = " OR "
				toSql += toEqul
			}
			toSql += ")"
			sql += toSql
		}
	} else {
		// multi chain
		sql = "select COUNT(DISTINCT from) from ("
		s1 := ""
		for k, v := range contracts {
			tblName, _ := getTableName(k)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			toSql := " AND ("
			s2 := ""
			for _, c := range v {
				toEqul := fmt.Sprintf(" %s (to = '%s')", s2, c)
				s2 = " OR "
				toSql += toEqul
			}
			toSql += ")"
			unionSql += toSql
			sql += s1 + unionSql
			s1 = " UNION ALL "
		}
		sql += " )"
	}
	//	log.Info("sql:%s", sql)
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	} else {
		log.Info("try update dau cache")
		if min, max, err := svr.getBlockScope(chains); err == nil {
			if (req.Start > min.Unix()) && (max.Unix() > req.End) {
				// only here update cache
				svr.dbCache.updateDau(key, count)
			}
		}
	}

	rsp = &pb.DauRsp{
		Dau: count,
	}
	err = nil
	return
}

func (svr *Server) ChainDau(ctx context.Context, req *pb.ChainGameReq) (rsp *pb.DauRsp, err error) {
	var count uint64
	sql := ""

	if len(req.Chains) == 0 {
		rsp = &pb.DauRsp{
			Dau: 0,
		}
		return
	} else if len(req.Chains) == 1 {
		// only one chain
		for _, c := range req.Chains {
			tblName, _ := getTableName(c)
			sql = fmt.Sprintf("SELECT countDistinct(from) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		}
	} else {
		// multi chain
		sql = "select COUNT(DISTINCT from) from ("
		s := ""
		for _, c := range req.Chains {
			tblName, _ := getTableName(c)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			sql += s + unionSql
			s = " UNION ALL "
		}
		sql += " )"
	}

	//log.Info("sql:%s", sql)

	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.DauRsp{
		Dau: count,
	}
	return
}

func (svr *Server) TxCount(ctx context.Context, req *pb.GameReq) (rsp *pb.TxCountRsp, err error) {
	var count uint64
	key := fmt.Sprintf("%v-%v", req.Start, req.End)
	for _, c := range req.Contracts {
		key += fmt.Sprintf("%v:%v", c.Chain, c.Address)
	}
	keyMd5 := md5.Sum([]byte(key))
	key = hex.EncodeToString(keyMd5[:])

	if count, err = svr.dbCache.getTxCount(key); err == nil {
		rsp = &pb.TxCountRsp{
			Count: count,
		}
		err = nil
		return
	}

	sql := ""
	var contracts map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	var chains []pb.Chain
	for _, c := range req.Contracts {
		contracts[c.Chain] = append(contracts[c.Chain], c.Address)
	}
	for k := range contracts {
		chains = append(chains, k)
	}

	if len(contracts) == 1 {
		// only one chain
		for k, v := range contracts {
			tblName, _ := getTableName(k)
			sql = fmt.Sprintf("SELECT COUNT(*) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			toSql := " AND ("
			s := ""
			for _, c := range v {
				toEqul := fmt.Sprintf(" %s (to = '%s')", s, c)
				s = " OR "
				toSql += toEqul
			}
			toSql += ")"
			sql += toSql
		}
	} else {
		// multi chain
		sql = "select COUNT(*) from ("
		s1 := ""
		for k, v := range contracts {
			tblName, _ := getTableName(k)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			toSql := " AND ("
			s2 := ""
			for _, c := range v {
				toEqul := fmt.Sprintf(" %s (to = '%s')", s2, c)
				s2 = " OR "
				toSql += toEqul
			}
			toSql += ")"
			unionSql += toSql
			sql += s1 + unionSql
			s1 = " UNION ALL "
		}
		sql += " )"
	}

	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	} else {
		if min, max, err := svr.getBlockScope(chains); err == nil {
			if (req.Start > min.Unix()) && (max.Unix() > req.End) {
				// only here update cache
				svr.dbCache.updateTxCount(key, count)
			}
		}
	}

	rsp = &pb.TxCountRsp{
		Count: count,
	}
	err = nil
	return
}

func (svr *Server) ChainTxCount(ctx context.Context, req *pb.ChainGameReq) (rsp *pb.TxCountRsp, err error) {
	sql := ""
	if len(req.Chains) == 0 {
		rsp = &pb.TxCountRsp{
			Count: 0,
		}
		return
	} else if len(req.Chains) == 1 {
		// only one chain
		for _, c := range req.Chains {
			tblName, _ := getTableName(c)
			sql = fmt.Sprintf("SELECT COUNT(*) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		}
	} else {
		// multi chain
		sql = "select COUNT(*) from ("
		s := ""
		for _, c := range req.Chains {
			tblName, _ := getTableName(c)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			sql += s + unionSql
			s = " UNION ALL "
		}
		sql += " )"
	}

	//log.Info("sql:%s", sql)

	var count uint64
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.TxCountRsp{
		Count: count,
	}
	return
}
