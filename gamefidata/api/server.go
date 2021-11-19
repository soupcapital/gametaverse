package api

import (
	"context"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Server struct {
	ctx       context.Context
	cancelFun context.CancelFunc
	dbClient  *mongo.Client
	opts      options
	httpd     http.Server
	router    *Router
	db        *mongo.Database
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

	svr.httpd = http.Server{
		Addr:           svr.opts.ListenAddr,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        svr,
	}
	svr.router = NewRouter()

	svr.initHandler()

	return
}

func (svr *Server) initHandler() {
	svr.router.RegistRaw("/gamefidata/api/v1/dau", &DAUHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/trx", &TrxHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/info", &InfoHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/sort", &SortHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/all", &AllHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/games", &GamesHandler{URLHdl{server: svr}})
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

	svr.db = svr.dbClient.Database(db.DBName)
	if svr.db == nil {
		log.Error("db is null, please init db first")
		return
	}
	return
}

func (svr *Server) Run() (err error) {
	err = svr.httpd.ListenAndServe()
	if err != nil {
		log.Error("ListenAndServe Error:%s", err.Error())
	}
	return
}

func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svr.router.DealRaw(r.URL.Path, w, r)
}
