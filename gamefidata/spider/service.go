package spider

import (
	"context"
	"sync"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Service struct {
	opts     options
	beacon   chan struct{}
	wg       sync.WaitGroup
	forward  *Spider
	backward *Spider
	ctx      context.Context
	cancel   context.CancelFunc

	monitorField string
	dbClient     *mongo.Client
	db           *mongo.Database
	monitorTbl   *mongo.Collection
	gameTbl      *mongo.Collection

	latestTS uint64
}

func New() *Service {
	s := &Service{
		beacon: make(chan struct{}),
	}
	return s
}

func (s *Service) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	s.ctx = context.Background()

	s.monitorField = db.MonitorFieldName + "_" + s.opts.Chain

	if err = s.initDB(s.opts.MongoURI); err != nil {
		log.Error("init db error: %s", err.Error())
		return
	}

	s.forward = NewSpider(s.opts, false)
	err = s.forward.Init()
	if err != nil {
		log.Error("Init forward spider error:%s", err.Error())
		return err
	}

	s.backward = NewSpider(s.opts, true)
	err = s.backward.Init()
	if err != nil {
		log.Error("Init backward spider error:%s", err.Error())
		return err
	}

	return err
}

func (s *Service) initDB(URI string) (err error) {
	s.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(s.ctx, 10*time.Second)
	err = s.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = s.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	s.db = s.dbClient.Database(db.DBName)
	if s.db == nil {
		log.Error("db %s is null, please init db first", db.DBName)
		return
	}

	s.monitorTbl = s.db.Collection(db.MonitorTableName)
	if s.monitorTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}

	s.gameTbl = s.db.Collection(db.GameTableName)
	if s.gameTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}
	return
}

func (s *Service) routine(ctx context.Context, sp *Spider) {
	s.wg.Add(1)
	go func() {
		sp.Run(ctx, s.beacon, &s.wg)
	}()
}

// this method must be invoke after s.cancel
func (s *Service) updateGames() (err error) {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	cursor, err := s.gameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var dbgames []db.Game
	if err = cursor.All(ctx, &dbgames); err != nil {
		log.Error("Find game error: ", err.Error())
		return
	}

	var games []*GameInfo
	for _, game := range dbgames {
		info := &GameInfo{
			Name:      game.Name,
			ID:        game.ID,
			Contracts: game.Contracts,
		}
		games = append(games, info)
	}
	s.backward.UpdateGames(games)
	s.forward.UpdateGames(games)
	return
}

func (s *Service) restart() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	if err := s.updateGames(); err != nil {
		log.Error("update games error:%s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.cancel = cancel
	s.routine(ctx, s.forward)
	s.routine(ctx, s.backward)
}

func (s *Service) checkConfig() (updated bool, err error) {
	log.Info("check config")
	ts, err := s.loadLatestTS()
	if err != nil {
		return false, err
	}
	if ts > s.latestTS { // if ts == 0 and s.latestTS == 0 , do nothing
		s.latestTS = ts
		return true, nil
	}
	return false, nil
}

func (s *Service) loadLatestTS() (ts uint64, err error) {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{
		"_id": s.monitorField,
	}
	curs, err := s.monitorTbl.Find(ctx, filter)
	if err != nil {
		log.Error("Find monitor error:", err.Error())
		return
	}

	for curs.Next(ctx) {
		var m db.Monitor
		curs.Decode(&m)
		log.Info("t_monitor:%v", m)
		ts = m.LatestUpdate
		break
	}
	return
}

func (s *Service) checkAndStart() {
	if updated, err := s.checkConfig(); err != nil {
		if updated {
			s.restart()
		}
	} else {
		log.Info("check config error:%s", err.Error())
	}
}

func (s *Service) Run() (err error) {
	s.checkAndStart()
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		s.checkAndStart()
	}
	return
}
