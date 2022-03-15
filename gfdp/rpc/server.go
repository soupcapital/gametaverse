package rpc

import (
	"context"
	"fmt"
	"net"

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
}

func NewServer() (svr *Server) {
	svr = &Server{}
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
			"max_execution_time": 60,
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

func (svr *Server) chain(chain pb.Chain) string {
	switch chain {
	case pb.Chain_BSC:
		return "bsc"
	case pb.Chain_ETH:
		return "eth"
	case pb.Chain_POLYGON:
		return "polygon"
	default:
		return "unknown"
	}
}

func (svr *Server) Dau(ctx context.Context, req *pb.GameReq) (rsp *pb.DauRsp, err error) {
	sql := ""

	//TODO: normal case
	var contracts map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	for _, c := range req.Contracts {
		contracts[c.Chain] = append(contracts[c.Chain], c.Address)
	}

	if len(contracts) == 1 {
		// only one chain
		for k, v := range contracts {
			chain := svr.chain(k)
			tblName := fmt.Sprintf("t_tx_%s", chain)
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
			chain := svr.chain(k)
			tblName := fmt.Sprintf("t_tx_%s", chain)
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

	log.Info("sql:%s", sql)

	var count uint64
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.DauRsp{
		Dau: count,
	}
	return
}

func (svr *Server) ChainDau(ctx context.Context, req *pb.ChainGameReq) (rsp *pb.DauRsp, err error) {
	sql := ""

	if len(req.Chains) == 0 {
		rsp = &pb.DauRsp{
			Dau: 0,
		}
		return
	} else if len(req.Chains) == 1 {
		// only one chain
		for _, c := range req.Chains {
			chain := svr.chain(c)
			tblName := fmt.Sprintf("t_tx_%s", chain)
			sql = fmt.Sprintf("SELECT countDistinct(from) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		}
	} else {
		// multi chain
		sql = "select COUNT(DISTINCT from) from ("
		s := ""
		for _, c := range req.Chains {
			chain := svr.chain(c)
			tblName := fmt.Sprintf("t_tx_%s", chain)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			sql += s + unionSql
			s = " UNION ALL "
		}
		sql += " )"
	}

	log.Info("sql:%s", sql)

	var count uint64
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.DauRsp{
		Dau: count,
	}
	return
}

func (svr *Server) TxCount(ctx context.Context, req *pb.GameReq) (rsp *pb.TxCountRsp, err error) {
	sql := ""
	//TODO: normal case
	var contracts map[pb.Chain][]string = make(map[pb.Chain][]string, 10)
	for _, c := range req.Contracts {
		contracts[c.Chain] = append(contracts[c.Chain], c.Address)
	}

	if len(contracts) == 1 {
		// only one chain
		for k, v := range contracts {
			chain := svr.chain(k)
			tblName := fmt.Sprintf("t_tx_%s", chain)
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
			chain := svr.chain(k)
			tblName := fmt.Sprintf("t_tx_%s", chain)
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

	log.Info("sql:%s", sql)

	var count uint64
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.TxCountRsp{
		Count: count,
	}
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
			chain := svr.chain(c)
			tblName := fmt.Sprintf("t_tx_%s", chain)
			sql = fmt.Sprintf("SELECT COUNT(*) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
		}
	} else {
		// multi chain
		sql = "select COUNT(*) from ("
		s := ""
		for _, c := range req.Chains {
			chain := svr.chain(c)
			tblName := fmt.Sprintf("t_tx_%s", chain)
			unionSql := fmt.Sprintf("SELECT  * from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
			sql += s + unionSql
			s = " UNION ALL "
		}
		sql += " )"
	}

	log.Info("sql:%s", sql)

	var count uint64
	if err := svr.dbConn.QueryRow(ctx, sql).Scan(&count); err != nil {
		log.Error("QueryRow error:%s", err.Error())
	}

	rsp = &pb.TxCountRsp{
		Count: count,
	}
	return
}
