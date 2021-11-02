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

type SortHandler struct {
	URLHdl
}

//Post is POST
func (hdl *SortHandler) Post(w http.ResponseWriter, r *http.Request) {
}

type GameDAU struct {
	GameID string `bson:"_id" json:"game"`
	DAU    int    `bson:"dau" json:"dau"`
}

type GameTrx struct {
	GameID string `bson:"_id" json:"game"`
	Trx    int    `bson:"trx" json:"count"`
}

//Get is GET
func (hdl *SortHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with sort")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	start := r.FormValue("start")
	end := r.FormValue("end")

	if len(end) == 0 ||
		len(start) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	startTime, err := time.Parse(cDateFormat, start)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}
	startTS := startTime.Unix()

	endTime, err := time.Parse(cDateFormat, end)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}
	endTS := endTime.Unix()

	log.Info("startDate:%v endDate:%v", startTS, endTS)

	if startTS > endTS {
		encoder.Encode(ErrParam)
		return
	}

	if startTS%cSecondofDay != 0 ||
		endTS%cSecondofDay != 0 {
		encoder.Encode(ErrTimestamp)
		return
	}
	gameTbl := hdl.server.db.Collection(db.DAUTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 100*time.Second)
	defer cancel()

	daus, err := hdl.dauByDate(ctx, gameTbl, startTS, endTS)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("daus: %s", err.Error())
		return
	}

	trxTbl := hdl.server.db.Collection(db.CountTableName)
	trxes, err := hdl.trxByDate(ctx, trxTbl, startTS, endTS)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("trxes: %s", err.Error())
		return
	}

	type Response struct {
		DAU []GameDAU `json:"dau"`
		Trx []GameTrx `json:"trx"`
	}

	rsp := Response{
		DAU: daus,
		Trx: trxes,
	}
	encoder.Encode(rsp)
}

func (hdl *SortHandler) dauByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (games []GameDAU, err error) {
	log.Info("games start:%v end:%v", start, end)
	groupStage := bson.M{
		"$group": bson.M{"_id": bson.M{
			"game": "$game",
			"user": "$user",
		}},
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
	groupStage2 := bson.M{
		"$group": bson.M{"_id": "$_id.game", "dau": bson.M{"$sum": 1}},
	}
	sortStage := bson.M{
		"$sort": bson.M{"dau": -1},
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, groupStage, groupStage2, sortStage)

	cur, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}

	for cur.Next(ctx) {
		rec := GameDAU{}
		cur.Decode(&rec)
		//log.Info("DAU aggregate record:%v", rec)
		games = append(games, rec)
	}
	return games, nil
}

func (hdl *SortHandler) trxByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (games []GameTrx, err error) {

	groupStage := bson.M{
		"$group": bson.M{"_id": "$game", "trx": bson.M{"$sum": "$count"}},
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
	sortStage := bson.M{
		"$sort": bson.M{"trx": -1},
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, groupStage, sortStage)

	cur, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}

	for cur.Next(ctx) {
		rec := GameTrx{}
		cur.Decode(&rec)
		//log.Info("Trx aggregate record:%v", rec)
		games = append(games, rec)
	}
	return games, nil
}
