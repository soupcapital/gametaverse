package api

import (
	"context"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Server struct {
	ctx       context.Context
	cancelFun context.CancelFunc
	dbClient  *mongo.Client
	opts      options
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

	err = svr.initDB(svr.opts.MongoURI)
	if err != nil {
		log.Error("Init mongon error:%s", err.Error())
		return err
	}

	return
}

func (svr *Server) initDB(URI string) (err error) {
	svr.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(svr.ctx, 10*time.Second)
	err = svr.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = svr.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	return
}

func (svr *Server) Run() (err error) {
	return
}
