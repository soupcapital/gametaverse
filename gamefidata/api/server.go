package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gfdp/rpc/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Server struct {
	ctx        context.Context
	cancelFun  context.CancelFunc
	dbClient   *mongo.Client
	opts       options
	httpd      http.Server
	router     *Router
	db         *mongo.Database
	monitorTbl *mongo.Collection
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
		ReadTimeout:    1000 * time.Second,
		WriteTimeout:   1000 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        svr,
	}
	svr.router = NewRouter()

	svr.initHandler()

	return
}

func (svr *Server) initHandler() {
	// svr.router.RegistRaw("/gamefidata/api/v1/dau", &DAUHandler{URLHdl{server: svr}})
	// svr.router.RegistRaw("/gamefidata/api/v1/trx", &TrxHandler{URLHdl{server: svr}})
	// svr.router.RegistRaw("/gamefidata/api/v1/all_chain", &AllChainHandler{URLHdl{server: svr}})
	// svr.router.RegistRaw("/gamefidata/api/v1/chain", &ChainHandler{URLHdl{server: svr}})

	svr.router.RegistRaw("/gamefidata/api/v1/info", &InfoHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/sort", &SortHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/all", &AllHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/total", &TotalHandlerV2{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/user2game", &User2GameHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/game2user", &Game2UserHandler{URLHdl{server: svr}})

	svr.router.RegistRaw("/gamefidata/api/v1/game_proj", &GameProjHandler{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/game_contract", &GameContractHandler{URLHdl{server: svr}})

	svr.router.RegistRaw("/gamefidata/api/v1/dau", &DauHandlerV2{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/trx", &TrxHandlerV2{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/all_chain", &AllChainHandlerV2{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/chain", &ChainHandlerV2{URLHdl{server: svr}})
	svr.router.RegistRaw("/gamefidata/api/v1/chain_hau", &ChainHauHandler{URLHdl{server: svr}})
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

	svr.monitorTbl = svr.db.Collection(db.MonitorTableName)
	if svr.monitorTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}
	return
}

func (svr *Server) UpdateMonitor(chain string) (err error) {
	ctx, cancel := context.WithTimeout(svr.ctx, 5*time.Second)
	defer cancel()

	monitorField := db.MonitorFieldName + "_" + chain
	opt := mngopts.Update()
	opt.SetUpsert(true)
	ts := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"latest": ts,
		},
	}
	_, err = svr.monitorTbl.UpdateByID(ctx, monitorField, update, opt)
	if err != nil {
		log.Error("Update  monitor latest error: ", err.Error())
		return
	}
	log.Info("Update  monitor latest:%d ", ts)
	return
}

func (svr *Server) getContracts(game string) (contracts []*pb.Contract, err error) {
	gameTbl := svr.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(svr.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{
		"id": game,
	})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	for _, g := range games {
		cn := ParsePBChain(g.Chain)
		for _, c := range g.Contracts {
			pc := &pb.Contract{
				Chain:   cn,
				Address: strings.ToLower(c),
			}
			contracts = append(contracts, pc)
		}
	}
	return
}

func (svr *Server) getAllGames() (games []db.Game, err error) {
	gameTbl := svr.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(svr.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	return
}

func (svr *Server) getAllContracts() (contracts []*pb.Contract, err error) {
	gameTbl := svr.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(svr.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	for _, g := range games {
		cn := ParsePBChain(g.Chain)
		if cn == pb.Chain_UNKNOWN {
			continue
		}
		for _, c := range g.Contracts {
			pc := &pb.Contract{
				Chain:   cn,
				Address: strings.ToLower(c),
			}
			contracts = append(contracts, pc)
		}
	}
	return
}

func (svr *Server) getGameContracts(id string) (contracts []*pb.Contract, err error) {
	gameTbl := svr.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(svr.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{
		"id": id,
	})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	for _, g := range games {
		cn := ParsePBChain(g.Chain)
		for _, c := range g.Contracts {
			pc := &pb.Contract{
				Chain:   cn,
				Address: strings.ToLower(c),
			}
			contracts = append(contracts, pc)
		}
	}
	return
}

func (svr *Server) getAllContractsOfChain(chain string) (contracts []*pb.Contract, err error) {
	if chain == "unknown" {
		return
	}
	gameTbl := svr.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(svr.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{
		"chain": chain,
	})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	for _, g := range games {
		cn := ParsePBChain(g.Chain)
		for _, c := range g.Contracts {
			pc := &pb.Contract{
				Chain:   cn,
				Address: strings.ToLower(c),
			}
			contracts = append(contracts, pc)
		}
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
