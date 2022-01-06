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

type AllHandler struct {
	URLHdl
}

//Delete is DELETE
func (hdl *AllHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Post is POST
func (hdl *AllHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *AllHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	dau, err := hdl.dauByDate(ctx, gameTbl, startTS, endTS)
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
		DAU int `json:"dau"`
		Trx int `json:"trx"`
	}

	rsp := Response{
		DAU: dau,
		Trx: trxes,
	}
	encoder.Encode(rsp)
}

func (hdl *AllHandler) dauByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (dau int, err error) {
	log.Info("games start:%v end:%v", start, end)
	groupStage := bson.M{
		"$group": bson.M{"_id": "$user"},
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
	countStage := bson.M{
		"$count": "dau",
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, groupStage, countStage)

	cur, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}

	for cur.Next(ctx) {
		rec := struct {
			DAU int `bson:"dau"`
		}{}
		cur.Decode(&rec)
		log.Info("DAU aggregate record:%v", rec)
		dau = rec.DAU
	}
	return dau, nil
}

func (hdl *AllHandler) trxByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (count int, err error) {

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
		count += rec.Trx
	}
	return count, nil
}
