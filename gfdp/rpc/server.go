package rpc

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
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

func (svr *Server) initDB() (err error) {
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
	return nil
}

func (svr *Server) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&svr.opts)
	}

	svr.ctx, svr.cancelFun = context.WithCancel(context.Background())
	log.Info("opts:%v", svr.opts)
	if err = svr.initDB(); err != nil {
		log.Error("Init DB error")
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
		if strings.Contains(err.Error(), "acquire conn timeout") {
			svr.initDB()
		}
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
		if strings.Contains(err.Error(), "acquire conn timeout") {
			svr.initDB()
		}
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

func (svr *Server) AllUserPrograms(ctx context.Context, req *pb.AllUserProgramsReq) (rsp *pb.AllUserProgramsRsp, err error) {
	rsp = &pb.AllUserProgramsRsp{}
	if len(req.Users) == 0 {
		return
	}

	sql := ""
	var contracts map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	for _, u := range req.Users {
		contracts[u.Chain] = append(contracts[u.Chain], u.Address)
	}
	user := req.Users[0].Address
	if len(contracts) == 1 {
		// only one chain
		for k := range contracts {
			tblName, _ := getTableName(k)
			sql = fmt.Sprintf("SELECT DISTINCT to  FROM %s WHERE from='%s' AND (ts > %d) AND (ts < %d)", tblName, user, req.Start, req.End)
		}
	} else {
		// multi chain
		sql = "select DISTINCT to  from ("
		s1 := ""
		for k := range contracts {
			tblName, _ := getTableName(k)
			unionSql := fmt.Sprintf("SELECT  DISTINCT to  from %s WHERE from='%s' AND (ts > %d) AND (ts < %d)", tblName, user, req.Start, req.End)
			sql += s1 + unionSql
			s1 = " UNION ALL "
		}
		sql += " )"
	}
	var tos []struct {
		To string `ch:"to"`
	}
	if err := svr.dbConn.Select(ctx, &tos, sql); err != nil {
		log.Error("QueryRow error:%s", err.Error())
		if strings.Contains(err.Error(), "acquire conn timeout") {
			svr.initDB()
		}
	}

	for _, to := range tos {
		rsp.Programs = append(rsp.Programs, to.To)
	}
	err = nil
	return
}

func (svr *Server) TwoGamesPlayers(ctx context.Context, req *pb.TwoGamesPlayersReq) (rsp *pb.TwoGamesPlayersRsp, err error) {
	rsp = &pb.TwoGamesPlayersRsp{}
	if len(req.GameOne) == 0 ||
		len(req.GameTwo) == 0 {
		return
	}

	sql := ""
	var contractsGameOne map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	for _, u := range req.GameOne {
		contractsGameOne[u.Chain] = append(contractsGameOne[u.Chain], u.Address)
	}

	var contractsGameTwo map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	for _, u := range req.GameTwo {
		contractsGameTwo[u.Chain] = append(contractsGameTwo[u.Chain], u.Address)
	}

	sqlOne := ""
	s1 := ""
	for k, v := range contractsGameOne {
		tblName, _ := getTableName(k)
		unionSql := fmt.Sprintf("SELECT  DISTINCT  from  from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		toSql := " AND ("
		s2 := ""
		for _, c := range v {
			toEqul := fmt.Sprintf(" %s (to = '%s')", s2, c)
			s2 = " OR "
			toSql += toEqul
		}
		toSql += ")"
		unionSql += toSql
		sqlOne += s1 + unionSql
		s1 = " UNION ALL "
	}

	sqlTwo := ""
	s1 = ""
	for k, v := range contractsGameTwo {
		tblName, _ := getTableName(k)
		unionSql := fmt.Sprintf("SELECT  DISTINCT  from  from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		toSql := " AND ("
		s2 := ""
		for _, c := range v {
			toEqul := fmt.Sprintf(" %s (to = '%s')", s2, c)
			s2 = " OR "
			toSql += toEqul
		}
		toSql += ")"
		unionSql += toSql
		sqlTwo += s1 + unionSql
		s1 = " UNION ALL "
	}

	sql = fmt.Sprintf(" (%s) INTERSECT (%s) ", sqlOne, sqlTwo)

	var tos []struct {
		From string `ch:"from"`
	}
	if err := svr.dbConn.Select(ctx, &tos, sql); err != nil {
		log.Error("QueryRow error:%s", err.Error())
		if strings.Contains(err.Error(), "acquire conn timeout") {
			svr.initDB()
		}
	}

	for _, to := range tos {
		rsp.Users = append(rsp.Users, to.From)
	}
	err = nil
	return
}
