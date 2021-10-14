package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DAUHandler struct {
	URLHdl
}

//Post is POST
func (hdl *DAUHandler) Post(w http.ResponseWriter, r *http.Request) {
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

	tableName := "t_" + game
	log.Info("tableName:%s", tableName)
	gameTbl := hdl.server.db.Collection(tableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()
	record := gameTbl.FindOne(ctx, bson.M{"timestamp": bson.M{"$gt": 0}})
	if record == nil {
		encoder.Encode(ErrGame)
		return
	}

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
		dau, err := hdl.dauByDate(ctx, gameTbl, theDay, theDay+cSecondofDay)
		if err != nil {
			encoder.Encode(ErrDB)
			log.Error("dauByDate: %s", err.Error())
			return
		}
		days = append(days, DayInfo{
			DAU:  dau,
			Date: theDay,
		})
		theDay += cSecondofDay
	}

	dau, err := hdl.dauByDate(ctx, gameTbl, startTS, endTS+cSecondofDay)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("dauByDate: %s", err.Error())
		return
	}

	type Response struct {
		Game  string    `json:"game"`
		Total int       `json:"total"`
		Data  []DayInfo `json:"data"`
	}

	rsp := Response{
		Game:  game,
		Data:  days,
		Total: dau,
	}
	encoder.Encode(rsp)
}

func (hdl *DAUHandler) dauByDate(ctx context.Context, gameTbl *mongo.Collection, start, end int64) (dau int, err error) {
	groupStage := bson.M{
		"$group": bson.M{"_id": "$from"},
	}

	matchStage1 := bson.M{
		"$match": bson.M{
			"timestamp": bson.M{"$gt": start},
		},
	}
	matchStage2 := bson.M{
		"$match": bson.M{
			"timestamp": bson.M{"$lt": end},
		},
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2, groupStage)

	curs, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}
	var transactions []bson.M
	err = curs.All(ctx, &transactions)
	if err != nil {
		log.Error(" curs.All error: %s", err.Error())
		return
	}
	log.Info("All:%d for %d", len(transactions), start)
	return len(transactions), nil
}
