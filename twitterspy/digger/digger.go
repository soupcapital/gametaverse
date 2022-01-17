package digger

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

func (d *Digger) makeScore(dayTS int64) (err error) {
	if dayTS <= twitterspy.SecOfDay {
		return nil
	}

	infos, err := d.queryDiggerInfoForOneDay(dayTS)
	if err != nil {
		log.Error("queryDiggerInfoForOneDay error:%s", err.Error())
		return
	}
	dayOneTs := dayTS - twitterspy.SecOfDay
	if err = d.calculateX(infos); err != nil {
		return
	}

	if err = d.calculateY(infos, dayOneTs); err != nil {
		return
	}

	if err = d.calculateZ(infos); err != nil {
		return
	}

	if err = d.updateScore(infos); err != nil {
		return
	}
	log.Info("============== dump[%v] =============", dayTS)
	d.dumpInfos(infos)

	return
}
func (d *Digger) dumpInfos(infos []*db.Digger) {
	//return
	for i, info := range infos {
		log.Info("%d:%v", i, info)
	}
}

func (d *Digger) calculateX(infos []*db.Digger) (err error) {
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].FavoriteCount >= infos[j].FavoriteCount
	})
	if infos[0].FavoriteCount != 0 {
		for _, info := range infos {
			info.Score = (10000) * 100 * float32(info.FavoriteCount) / float32(infos[0].FavoriteCount)
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].TweetsCount >= infos[j].TweetsCount
	})
	if infos[0].TweetsCount != 0 {
		for _, info := range infos {
			info.Score += 200 * float32(info.TweetsCount) / float32(infos[0].TweetsCount)
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ReplyCount >= infos[j].ReplyCount
	})
	if infos[0].ReplyCount != 0 {
		for _, info := range infos {
			info.Score += 300 * float32(info.ReplyCount) / float32(infos[0].ReplyCount)
		}
	}

	return
}

func (d *Digger) calculateY(infos []*db.Digger, dayOneTs int64) (err error) {
	dayOneinfos, err := d.queryDiggerInfoForOneDay(dayOneTs)
	if err != nil {
		log.Error("queryDiggerInfoForOneDay dayOne error:%s", err.Error())
		return
	}

	for _, info := range infos {
		found := false
		for _, dayOneInfo := range dayOneinfos {
			if info.Name == dayOneInfo.Name {
				found = true
				if dayOneInfo.FollowerCount != 0 {
					info.Score *= (float32(info.FollowerCount) - float32(dayOneInfo.FollowerCount)) / float32(dayOneInfo.FollowerCount)
				} else {
					info.Score *= 1
				}
			}
			if !found {
				info.Score *= 1
			}
		}
	}

	return
}

func (d *Digger) calculateZ(infos []*db.Digger) (err error) {
	for _, info := range infos {
		if info.TweetsCount == 0 {
			info.Score /= 99
		} else {
			info.Score /= float32(info.TweetsCount)
		}
	}
	return
}

func (d *Digger) updateScore(infos []*db.Digger) (err error) {
	diggerTbl := d.db.Collection(db.DiggerTable)
	for _, info := range infos {
		ctx, cancel := context.WithTimeout(d.ctx, 3*time.Second)
		defer cancel()
		id := info.ID

		opt := mngopts.Update()
		opt.SetUpsert(true)

		update := bson.M{
			"$set": bson.M{
				"score": info.Score,
			},
		}
		_, err = diggerTbl.UpdateByID(ctx, id, update, opt)
		if err != nil {
			log.Error("Update [%s] score error: ", info.ID, err.Error())
			return
		}
	}
	return
}

func (d *Digger) queryDiggerInfoForOneDay(dateTS int64) (infos []*db.Digger, err error) {
	ctx, cancel := context.WithTimeout(d.ctx, 1000*time.Second)
	defer cancel()

	diggerTbl := d.db.Collection(db.DiggerTable)
	curs, err := diggerTbl.Find(ctx, bson.M{"ts": dateTS})
	if err != nil {
		log.Error("Find ts[%v] error", dateTS)
		return
	}
	if err = curs.All(ctx, &infos); err != nil {
		log.Error("Decode digger infos error: ", err.Error())
		return
	}
	return
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

func Start() (err error) {
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
	return
}

func MakeScore(date string) (err error) {
	dayTime, err := time.Parse(twitterspy.DateFormat, date)
	if err != nil {
		return
	}
	dayTS := dayTime.Unix()
	if dayTS%twitterspy.SecOfDay != 0 {
		log.Error("dateTS:%v", dayTS)
		return errors.New("worng date")
	}
	log.Info("Deal score for ts:%v", dayTS)
	return _digger.makeScore(dayTS)
}
