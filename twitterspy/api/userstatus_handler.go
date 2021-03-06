package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
)

type UserStatusHandler struct {
	URLHdl
}

//Post is POST
func (hdl *UserStatusHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Post is DELETE
func (hdl *UserStatusHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *UserStatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vname := r.FormValue("vname")
	date := r.FormValue("date")

	if len(vname) == 0 ||
		len(date) == 0 {
		log.Error("vname:%s date:%s", vname, date)
		encoder.Encode(ErrParam)
		return
	}

	dayTime, err := time.Parse(twitterspy.DateFormat, date)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}
	dayTS := dayTime.Unix()

	if dayTS%twitterspy.SecOfDay != 0 {
		log.Error("dateTS:%v", dayTS)
		encoder.Encode(ErrTimestamp)
		return
	}

	if dayTS <= twitterspy.SecOfDay {
		encoder.Encode(ErrTimestamp)
		return
	}

	favCount := 0
	replyCount := 0
	retweetCount := 0

	dayOneTS := dayTS - twitterspy.SecOfDay
	dayOneInfo, err := hdl.queryDiggerInfoForOneDay(vname, dayOneTS)
	if err != nil {
		log.Error("queryDiggerInfoForOneDay day one info error:%s", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	dayInfo, err := hdl.queryDiggerInfoForOneDay(vname, dayTS)
	if err != nil {
		log.Error("queryDiggerInfoForOneDay day info error:%s", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	if dayOneInfo == nil || dayInfo == nil {
		log.Error("day one info is nil")
		encoder.Encode(ErrNoDataForDay)
		return
	}
	favCount = dayInfo.FavoriteCount
	replyCount = dayInfo.ReplyCount
	retweetCount = dayInfo.RetweetCount

	type Response struct {
		Increase      int     `increase`
		IncreaseRate  float32 `increase_rate`
		TweetCount    int     `tweet_count`
		ReplyCount    int     `tweet_reply_count`
		RetweetCount  int     `tweet_retweet_count`
		FavoriteCount int     `tweet_favorite_count`
		Score         float32 `score`
		Err           int     `json:"errno"`
		ErrMsg        string  `json:"errmsg"`
	}
	inc := dayInfo.FollowerCount - dayOneInfo.FollowerCount
	incRate := float32(0.0)
	if dayOneInfo.FollowerCount != 0 {
		incRate = float32(inc) / float32(dayOneInfo.FollowerCount)
	}
	rsp := Response{
		Increase:      inc,
		IncreaseRate:  float32(incRate),
		TweetCount:    dayInfo.TweetsCount - dayOneInfo.TweetsCount,
		FavoriteCount: favCount,
		ReplyCount:    replyCount,
		RetweetCount:  retweetCount,
		Score:         dayInfo.Score,
		Err:           0,
		ErrMsg:        "",
	}
	encoder.Encode(rsp)
}

func (hdl *UserStatusHandler) queryDiggerInfoForOneDay(name string, dateTS int64) (info *db.Digger, err error) {
	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	id := fmt.Sprintf("%s_%d", name, dateTS)
	diggerTbl := hdl.server.db.Collection(db.DiggerTable)
	sr := diggerTbl.FindOne(ctx, bson.M{"_id": id})
	if sr == nil || sr.Err() != nil {
		log.Error("Find vname[%s] error", id)
		return
	}
	userStatus := &db.Digger{}
	err = sr.Decode(userStatus)
	if err != nil {
		log.Error("bson decode error: %s", err.Error())
	}
	info = userStatus
	return
}
