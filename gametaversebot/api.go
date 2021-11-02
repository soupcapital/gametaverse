package gametaversebot

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gametaversebot/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type API struct {
	cli      *http.Client
	apiAddr  string
	ticker   *time.Ticker
	newsChan chan (*News)
	ctx      context.Context
	dbClient *mongo.Client
	db       *mongo.Database
	guardTbl *mongo.Collection
	oldNews  []int
}

func NewAPI() *API {
	api := &API{}
	return api
}

func (api *API) Init(addr string, dbAddr string, newsChan chan (*News)) (err error) {
	api.cli = &http.Client{}
	api.apiAddr = addr
	api.ticker = time.NewTicker(10 * time.Second)
	api.ctx = context.Background()
	api.newsChan = newsChan
	err = api.initDB(dbAddr)
	if err != nil {
		log.Error("init db error:%s", err.Error())
		return
	}
	err = api.loadGuard()
	if err != nil {
		log.Error("load guard error:%s", err.Error())
		return
	}
	return
}

func (api *API) loadGuard() (err error) {
	ctx, cancel := context.WithTimeout(api.ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{
		"_id": db.GuardFiledID,
	}
	curs, err := api.guardTbl.Find(ctx, filter)
	if err != nil {
		log.Error("Find monitor error:", err.Error())
		return err
	}

	for curs.Next(ctx) {
		var m db.Guard
		curs.Decode(&m)
		log.Info("t_guard:%v", m)
		api.oldNews = make([]int, len(m.News))
		copy(api.oldNews, m.News)
		break
	}

	log.Info("oldNews:%+v", api.oldNews)
	return
}

func (api *API) updateGuard(news []int) (err error) {
	ctx, cancel := context.WithTimeout(api.ctx, 5*time.Second)
	defer cancel()

	sort.Ints(news)

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"news": news,
		},
	}
	_, err = api.guardTbl.UpdateByID(ctx, db.GuardFiledID, update, opt)
	if err != nil {
		log.Error("Update top block error: ", err.Error())
		return
	}
	log.Info("Update guard to:%+v ", news)
	api.oldNews = make([]int, len(news))
	copy(api.oldNews, news)
	return
}

func (api *API) initDB(URI string) (err error) {
	api.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(api.ctx, 10*time.Second)
	err = api.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = api.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	api.db = api.dbClient.Database(db.DBName)
	if api.db == nil {
		log.Error("db solana-spl is null, please init db first")
		return
	}

	api.guardTbl = api.db.Collection(db.GuardTableName)
	if api.guardTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}

	return
}

func (api *API) Run() {
	err := api.loadGuard()
	if err != nil {
		log.Error("load guard error:%s", err.Error())
		return
	}
	for range api.ticker.C {
		news, err := api.QueryNews()
		if err != nil {
			log.Error("Query news error:%s", err.Error())
			continue
		}
		api.dealNews(news)
	}
}

func (api *API) dealNews(news []*News) {
	var guardNews []int
	for _, n := range news {
		guardNews = append(guardNews, n.ID)
		i := sort.SearchInts(api.oldNews, n.ID)

		if !(i < len(api.oldNews) && api.oldNews[i] == n.ID) {
			// new news
			api.newsChan <- n
		}

	}
	api.updateGuard(guardNews)
}

func (api *API) QueryNews() (news []*News, err error) {
	req, err := http.NewRequest("GET", api.apiAddr, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return
	}

	resp, err := api.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()

	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Newss []*News `json:"newss"`
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	news = make([]*News, len(respJOSN.Newss))
	copy(news, respJOSN.Newss)
	log.Info("news:%+v", news)
	return
}
