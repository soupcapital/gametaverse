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

func (svr *Server) Dau(ctx context.Context, req *pb.DauReq) (rsp *pb.DauRsp, err error) {
	chain := svr.chain(req.Chain)
	tblName := fmt.Sprintf("t_tx_%s", chain)

	sql := fmt.Sprintf("SELECT countDistinct(from) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
	if len(req.Contracts) > 0 {
		toSql := " AND ("
		s := ""
		for _, c := range req.Contracts {
			toEqul := fmt.Sprintf(" %s (to = '%s')", s, c)
			s = " OR "
			toSql += toEqul
		}
		toSql += ")"
		sql += toSql
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

func (svr *Server) TxCount(ctx context.Context, req *pb.TxCountReq) (rsp *pb.TxCountRsp, err error) {
	chain := svr.chain(req.Chain)
	tblName := fmt.Sprintf("t_tx_%s", chain)

	sql := fmt.Sprintf("SELECT COUNT(*) from %s WHERE (ts > %d) AND (ts < %d)", tblName, req.Start, req.End)
	if len(req.Contracts) > 0 {
		toSql := " AND ("
		s := ""
		for _, c := range req.Contracts {
			toEqul := fmt.Sprintf("%s (to = '%s') ", s, c)
			s = " OR "
			toSql += toEqul
		}
		toSql += " ) "
		sql += toSql
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
