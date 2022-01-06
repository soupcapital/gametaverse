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

type DAUHandler struct {
	URLHdl
}

//Post is POST
func (hdl *DAUHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Delete is DELETE
func (hdl *DAUHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *DAUHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	game := r.FormValue("gameid")
	start := r.FormValue("start")
	end := r.FormValue("end")

	if len(game) == 0 ||
		len(end) == 0 ||
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

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	theDay := startTS
	type DayInfo struct {
		Date int64 `json:"date"`
		DAU  int   `json:"dau"`
	}

	var days []DayInfo
	for {
		if theDay > endTS {
			break
		}
		dau, err := hdl.dauByDate(ctx, game, gameTbl, theDay, theDay+cSecondofDay)
		if err != nil {
			encoder.Encode(ErrDB)
			log.Error("dauByDate: %s", err.Error())
			return
		}
		log.Info("dau:%v for date:%v", dau, theDay)
		days = append(days, DayInfo{
			DAU:  dau,
			Date: theDay,
		})

		theDay += cSecondofDay
	}

	// dau, err := hdl.dauByDate(ctx, game, gameTbl, startTS, endTS+cSecondofDay)
	// if err != nil {
	// 	encoder.Encode(ErrDB)
	// 	log.Error("dauByDate: %s", err.Error())
	// 	return
	// }

	type Response struct {
		Game  string    `json:"game"`
		Total int       `json:"total"`
		Data  []DayInfo `json:"data"`
	}

	rsp := Response{
		Game: game,
		Data: days,
		//	Total: dau,
	}
	encoder.Encode(rsp)
}

func (hdl *DAUHandler) dauByDate(ctx context.Context, game string, gameTbl *mongo.Collection, start, end int64) (dau int, err error) {
	log.Info("game:%v start:%v end:%v", game, start, end)
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
			"ts": bson.M{"$lt": end},
		},
	}
	matchStage3 := bson.M{
		"$match": bson.M{
			"game": game,
		},
	}
	countStage := bson.M{
		"$count": "dau",
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, matchStage3, groupStage, countStage)

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
