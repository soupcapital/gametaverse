package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GamesHandler struct {
	URLHdl
}

//Post is POST
func (hdl *GamesHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *GamesHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with sort")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	endTime := time.Now()
	endTS := endTime.Unix()
	startTS := endTS - cSecondofDay*5

	log.Info("startDate:%v endDate:%v", startTS, endTS)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 100*time.Second)
	defer cancel()
	trxTbl := hdl.server.db.Collection(db.CountTableName)
	games, err := hdl.gamesByDate(ctx, trxTbl, startTS, endTS)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("trxes: %s", err.Error())
		return
	}

	type Response struct {
		Games []string `json:"games"`
	}
	rsp := Response{
		Games: games,
	}
	encoder.Encode(rsp)
}

func (hdl *GamesHandler) gamesByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (games []string, err error) {

	groupStage := bson.M{
		"$group": bson.M{"_id": "$game"},
	}
	matchStage1 := bson.M{
		"$match": bson.M{
			"ts": bson.M{"$gte": start},
		},
	}
	matchStage2 := bson.M{
		"$match": bson.M{
			"ts": bson.M{"$lte": end},
		},
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, groupStage)

	cur, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}

	for cur.Next(ctx) {
		rec := GameTrx{}
		cur.Decode(&rec)
		//log.Info("Trx aggregate record:%v", rec)
		//games = append(games, rec)
		games = append(games, rec.GameID)
	}
	err = nil
	return
}
