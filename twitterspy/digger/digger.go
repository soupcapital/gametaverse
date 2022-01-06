package digger

import (
	"context"
	"fmt"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Digger struct {
	ctx context.Context

	db       *mongo.Database
	dbClient *mongo.Client
	conn     *twitterspy.TwitterSearchConn
}

var _digger Digger

func (d *Digger) initDB(URI string) (err error) {
	d.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
	defer cancel()
	err = d.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = d.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	d.db = d.dbClient.Database(db.DBName)
	if d.db == nil {
		log.Error("db is null, please init db first")
		return
	}
	return
}

func (d *Digger) dealUser(name string) (err error) {
	log.Info("deal user:%v", name)

	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	userInfo, err := d.conn.QueryUserInfo(name)
	if err == twitterspy.ErrTokenForbid {
		if err = d.conn.RefreshToken(); err != nil {
			log.Error("Refresh token error:%s", err.Error())
			return
		}
		userInfo, err = d.conn.QueryUserInfo(name)
		if err != nil {
			log.Error("QueryUserInfo  error:%s", err.Error())
			return
		}
	}
	if err != nil {
		return
	}

	ts := time.Now().Unix() / twitterspy.SecOfDay * twitterspy.SecOfDay
	until := time.Unix(ts, 0)
	since := time.Unix(ts-twitterspy.SecOfDay, 0)

	tweets, err := d.conn.QueryV(name, since, until, 100)
	if err != nil {
		if err == twitterspy.ErrTokenForbid {
			if err = d.conn.RefreshToken(); err != nil {
				log.Error("Refresh token error:%s", err.Error())
				return
			}
			tweets, err = d.conn.QueryV(name, since, until, 100)
			if err != nil {
				log.Error("query tweets error:%s", err.Error())
				return
			}
		} else {
			log.Error("query tweets error:%s", err.Error())
			return
		}
	}

	favCount := 0
	replyCount := 0
	retweetCount := 0

	for _, tweet := range tweets {
		favCount += tweet.FavoriteCount
		replyCount += tweet.ReplyCount
		retweetCount += tweet.RetweetCount
	}

	diggerTbl := d.db.Collection(db.DiggerTable)
	id := fmt.Sprintf("%s_%d", name, ts)

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"name": name,
			"fc":   userInfo.Legacy.FollowersCount,
			"tc":   userInfo.Legacy.StatusesCount,
			"ftc":  favCount,
			"rpc":  replyCount,
			"rtc":  retweetCount,
			"ts":   ts,
		},
	}
	_, err = diggerTbl.UpdateByID(ctx, id, update, opt)
	if err != nil {
		log.Error("Update top block error: ", err.Error())
		return
	}

	return
}

func (d *Digger) tracedUsers() (traced []string, err error) {

	ctx, cancel := context.WithTimeout(d.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	vnameTbl := d.db.Collection(db.VNameTable)
	cursor, err := vnameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find vname error: ", err.Error())
		return
	}

	var vnames []db.VName
	if err = cursor.All(ctx, &vnames); err != nil {
		log.Error("Find vname error: ", err.Error())
		return
	}
	for _, vname := range vnames {
		if vname.Status == int8(db.VNSTraced) {
			traced = append(traced, vname.ID)
		}
	}
	return
}

func (d *Digger) loop() {
	log.Info("loop once")
	users, err := d.tracedUsers()
	if err != nil {
		log.Info("get traced users error")
		return
	}
	for _, user := range users {
		d.dealUser(user)
		time.Sleep(200 * time.Millisecond)
	}
}

func Init(mongoAddr string, tokenRPC string) (err error) {
	_digger.ctx = context.Background()

	_digger.conn = twitterspy.NewTwitterSearchConn()
	if err = _digger.conn.Init(tokenRPC); err != nil {
		return err
	}

	if err = _digger.initDB(mongoAddr); err != nil {
		return err
	}
	return
}

func Start() {
	done := make(chan (struct{}))
	nowTS := time.Now().Unix()
	dt := twitterspy.SecOfDay - (nowTS % twitterspy.SecOfDay)
	loopTimer := time.NewTimer(time.Duration(dt * int64(time.Second)))
	for {
		select {
		case <-loopTimer.C:
			_digger.loop()
			loopTimer.Reset(time.Duration(twitterspy.SecOfDay * int64(time.Second)))
		case <-done:
			return
		}
	}
}
