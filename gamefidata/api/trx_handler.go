package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"go.mongodb.org/mongo-driver/bson"
)

type TrxHandler struct {
	URLHdl
}

//Post is POST
func (hdl *TrxHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *TrxHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	game := r.FormValue("gameid")
	date := r.FormValue("date")

	// var err error
	// var startTS uint64
	// var amount float64

	if len(game) == 0 ||
		len(date) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	startTime, _ := time.Parse(cDateFormat, date)
	log.Info("t is %d", startTime.Unix())

	tableName := "t_" + game
	log.Info("tableName:%s", tableName)
	gameTbl := hdl.server.db.Collection(tableName)

	ctx, _ := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	record := gameTbl.FindOne(ctx, bson.M{"timestamp": bson.M{"$gt": 0}})
	if record == nil {
		encoder.Encode(ErrGame)
		return
	}
	matchStage1 := bson.M{
		"$match": bson.M{
			"timestamp": bson.M{"$gt": startTime.Unix()},
		},
	}
	matchStage2 := bson.M{
		"$match": bson.M{
			"timestamp": bson.M{"$lt": startTime.Unix() + cSecondofDay},
		},
	}

	pipeline := []bson.M{}
	pipeline = append(pipeline, matchStage1, matchStage2)

	curs, err := gameTbl.Aggregate(ctx, pipeline)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("Aggregate error: %s", err.Error())
		return
	}
	var transactions []bson.M
	err = curs.All(ctx, &transactions)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error(" curs.All error: %s", err.Error())
		return
	}
	log.Info("All:%d", len(transactions))

	type Response struct {
		Game     string `json:"game"`
		TrxCount int    `json:"trx_count"`
	}
	rsp := Response{
		Game:     game,
		TrxCount: len(transactions),
	}
	encoder.Encode(rsp)
}
