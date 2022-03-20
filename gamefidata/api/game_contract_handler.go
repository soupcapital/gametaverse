package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/spider"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
)

type GameContractHandler struct {
	URLHdl
}

//Post is POST
func (hdl *GameContractHandler) Post(w http.ResponseWriter, r *http.Request) {

	type PostRequest struct {
		GameID    string   `json:"game_id"`
		Chain     string   `json:"chain"`
		Contracts []string `json:"contracts"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	var req PostRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Error("Decode request error:%s", err.Error())
		encoder.Encode(ErrParam)
		return
	}
	if len(req.GameID) == 0 ||
		len(req.Chain) == 0 {
		encoder.Encode(ErrParam)
		return
	}
	if !spider.ValiedChainName(req.Chain) {
		encoder.Encode(ErrUnknownChain)
		return
	}

	if existed, err := hdl.isGameExisted(req.GameID, req.Chain); !existed {
		if err != nil {
			log.Error("game:%s isGameExisted error:%s", req.GameID, err.Error())
			encoder.Encode(ErrDB)
			return
		}
		encoder.Encode(ErrNoSuchGame)
		return
	}

	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"contracts": req.Contracts,
		},
	}
	//_, err = gameTbl.UpdateByID(ctx, req.GameID, update, opt)
	filter := bson.M{
		"_id": fmt.Sprintf("%s@%s", req.GameID, req.Chain),
	}
	_, err = gameTbl.UpdateMany(ctx, filter, update, opt)
	if err != nil {
		log.Error("Update contracts for game[%s] error: ", req.GameID, err.Error())
		encoder.Encode(ErrDB)
		return
	}

	if err = hdl.server.UpdateMonitor(req.Chain); err != nil {
		log.Error("update monitor lastest error:%s", err.Error())
		encoder.Encode(ErrUpdateMonitor)
		return
	}

	type Response struct {
		GameID string `json:"game_id"`
		Err    int    `json:"errno"`
		ErrMsg string `json:"errmsg"`
	}

	rsp := Response{
		GameID: req.GameID,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}

//Delete is DELETE
func (hdl *GameContractHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *GameContractHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	gameID := r.FormValue("game_id")
	chain := r.FormValue("chain")

	if len(gameID) == 0 || len(chain) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": fmt.Sprintf("%s@%s", gameID, chain),
	}

	rst := gameTbl.FindOne(ctx, filter)
	if rst.Err() != nil {
		if rst.Err() == mongo.ErrNoDocuments {
			log.Error("get without game[%s]", gameID)
			encoder.Encode(ErrNoSuchGame)
			return
		}
		log.Error("Find vname error: ", rst.Err())
		encoder.Encode(ErrDB)
		return
	}

	var m db.Game
	if err := rst.Decode(&m); err != nil {
		log.Error("Decode game[%s] error:%s ", gameID, rst.Err())
		encoder.Encode(ErrDB)
		return
	}

	type Response struct {
		Game      string   `json:"game"`
		Contracts []string `json:"contracts"`
	}
	rsp := Response{
		Game:      gameID,
		Contracts: m.Contracts,
	}
	encoder.Encode(rsp)
}

func (hdl *GameContractHandler) isGameExisted(gameID string, chain string) (existed bool, err error) {

	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	filter := bson.M{
		"_id": fmt.Sprintf("%s@%s", gameID, chain),
	}

	rst := gameTbl.FindOne(ctx, filter)
	if rst.Err() != nil {
		if rst.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		log.Error("Find vname error: ", rst.Err())
		return false, rst.Err()
	}
	return true, nil
}
