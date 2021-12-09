package twitterspy

import (
	"context"
	"fmt"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type TwitterSpider struct {
	tgbot    *TGBot
	token    *Token
	vs       []string
	conn     *TwitterSearchConn
	internal time.Duration
	perCount uint32
	msgChan  chan (TweetInfo)
	ctx      context.Context
	db       *mongo.Database
	dbClient *mongo.Client
	vtable   *mongo.Collection
}

func NewTwitterSpider() *TwitterSpider {
	ts := &TwitterSpider{}
	ts.token = NewToken()
	ts.conn = NewTwitterSearchConn()
	ts.ctx = context.Background()
	return ts
}

func (ts *TwitterSpider) initDB(uri string) (err error) {
	ts.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(uri))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(ts.ctx, 10*time.Second)
	err = ts.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = ts.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	ts.db = ts.dbClient.Database(db.DBName)
	if ts.db == nil {
		log.Error("db is null, please init db first")
		return
	}

	ts.vtable = ts.db.Collection(db.VNameTable)
	return
}

func (ts *TwitterSpider) Init(msgChan chan (TweetInfo), vs []string, internal time.Duration, count uint32, dbUrl string) (err error) {
	_vs := make([]string, len(vs))
	copy(_vs, vs)
	ts.vs = _vs

	if err = ts.token.Refresh(); err != nil {
		return err
	}
	if err = ts.conn.Init(ts.token.token); err != nil {
		return err
	}
	ts.internal = internal
	ts.perCount = count
	ts.msgChan = msgChan

	if err = ts.initDB(dbUrl); err != nil {
		return err
	}
	return nil
}

func (ts *TwitterSpider) Start() (err error) {
	ticker := time.NewTicker(ts.internal)
	ts.updateTwitter()
	for {
		select {
		case <-ticker.C:
			ts.updateTwitter()
		}
	}
}

func (ts *TwitterSpider) updateTwitter() {
	for _, v := range ts.vs {
	AGAIN:
		tweets, err := ts.conn.QueryV(v, ts.internal, ts.perCount)
		if err != nil {
			if err == ErrTokenForbid {
				if err = ts.token.Refresh(); err == nil {
					ts.conn.token = ts.token.token
					log.Info("Refresh token success and goto Again")
					goto AGAIN
				}
			}
			log.Error("QueryV error:%s", err.Error())
			continue
		}
		log.Info("Query %v Got :%v", v, tweets)
		for _, t := range tweets {
			t.Author = v
			msg := fmt.Sprintf("[%s@%s]:%s", v, time.Time(t.CreateAt).String(), t.FullText)
			log.Info("TWEET:%s", msg)
			ts.msgChan <- t
		}
		time.Sleep(100 * time.Millisecond)
	}
}
