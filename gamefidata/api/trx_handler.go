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

type TrxHandler struct {
	URLHdl
}

//Post is POST
func (hdl *TrxHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *TrxHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with trx")
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

	gameTbl := hdl.server.db.Collection(db.CountTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	type DayInfo struct {
		Date  int64 `json:"date"`
		Count int   `json:"count"`
	}
	var days []DayInfo
	theDay := startTS
	total := 0
	for {
		if theDay > endTS {
			break
		}
		count, err := hdl.trxByDate(ctx, game, gameTbl, theDay)
		if err != nil {
			encoder.Encode(ErrDB)
			log.Error("trxByDate: %s", err.Error())
			return
		}
		days = append(days, DayInfo{
			Count: count,
			Date:  theDay,
		})
		total += count
		theDay += cSecondofDay
	}

	type Response struct {
		Game  string    `json:"game"`
		Data  []DayInfo `json:"data"`
		Total int       `json:"total"`
	}
	rsp := Response{
		Game:  game,
		Data:  days,
		Total: total,
	}
	encoder.Encode(rsp)
}

func (hdl *TrxHandler) trxByDate(ctx context.Context, game string, gameTbl *mongo.Collection, start int64) (count int, err error) {

	filter := bson.M{
		"game": game,
		"ts":   start,
	}

	cur, err := gameTbl.Find(ctx, filter)
	if err != nil {
		log.Error("Aggregate error: %s", err.Error())
		return
	}

	for cur.Next(ctx) {
		rec := db.Count{}
		cur.Decode(&rec)
		log.Info("Trx aggregate record:%v", rec)
		count = int(rec.Count)
	}
	return count, nil
}
