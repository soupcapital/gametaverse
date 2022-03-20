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

type GameProjHandler struct {
	URLHdl
}

//Post is POST
func (hdl *GameProjHandler) Post(w http.ResponseWriter, r *http.Request) {
	type PostRequest struct {
		GameID    string   `json:"game_id"`
		GameName  string   `json:"game_name"`
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
		len(req.GameName) == 0 ||
		len(req.Chain) == 0 {
		encoder.Encode(ErrParam)
		return
	}
	if !spider.ValiedChainName(req.Chain) {
		encoder.Encode(ErrUnknownChain)
		return
	}

	if existed, err := hdl.isGameExisted(req.GameID, req.Chain); existed {
		if err != nil {
			log.Error("game:%s isGameExisted error:%s", req.GameID, err.Error())
			encoder.Encode(ErrDB)
			return
		}
		encoder.Encode(ErrGameExisted)
		return
	}

	if err := hdl.insertGame(req.GameID, req.GameName, req.Chain, req.Contracts); err != nil {
		log.Error("Insert game:%s error:%s", req.GameID, err.Error())
		encoder.Encode(ErrInsertGame)
		return
	}

	type Response struct {
		GameID   string `json:"game_id"`
		GameName string `json:"game_name"`
		Err      int    `json:"errno"`
		ErrMsg   string `json:"errmsg"`
	}

	rsp := Response{
		GameID:   req.GameID,
		GameName: req.GameName,
		Err:      0,
		ErrMsg:   "",
	}
	encoder.Encode(rsp)

}

//Delete is DELETE
func (hdl *GameProjHandler) Delete(w http.ResponseWriter, r *http.Request) {
	type PostRequest struct {
		GameID string `json:"game_id"`
		Chain  string `json:"chain"`
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

	if err := hdl.deleteGame(req.GameID, req.Chain); err != nil {
		log.Error("delete game %s error:%s", req.GameID, err.Error())
		encoder.Encode(ErrDeleteGame)
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

//Get is GET
func (hdl *GameProjHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	cursor, err := gameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find game error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Error("Find game error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	type RspGameInfo struct {
		GameID   string `json:"game_id"`
		GameName string `json:"game_name"`
		Chain    string `json:"chain"`
	}
	var gameInfos []*RspGameInfo
	for _, game := range games {
		gameInfo := &RspGameInfo{
			GameID:   game.GameID,
			GameName: game.Name,
			Chain:    game.Chain,
		}
		gameInfos = append(gameInfos, gameInfo)
	}

	type Response struct {
		Games []*RspGameInfo `json:"games"`
	}

	rsp := Response{
		Games: gameInfos,
	}
	encoder.Encode(rsp)
}

func (hdl *GameProjHandler) isGameExisted(gameID string, chain string) (existed bool, err error) {

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

func (hdl *GameProjHandler) insertGame(gameID string, gameName string, chain string, contracts []string) (err error) {

	game := &db.Game{
		ID:        fmt.Sprintf("%s@%s", gameID, chain),
		GameID:    gameID,
		Name:      gameName,
		Chain:     chain,
		Contracts: contracts,
	}

	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	rst, err := gameTbl.InsertOne(ctx, game)
	if err != nil {
		log.Error("insert game:%s error:%s", gameID, err.Error())
	}
	log.Info("insert game[%v] success with:%v", game, rst.InsertedID)
	if err = hdl.server.UpdateMonitor(chain); err != nil {
		log.Error("update monitor lastest error:%s", err.Error())
		return
	}
	return
}

func (hdl *GameProjHandler) deleteGame(gameID, chain string) (err error) {
	gameTbl := hdl.server.db.Collection(db.GameTableName)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": fmt.Sprintf("%s@%s", gameID, chain),
	}

	rst, err := gameTbl.DeleteMany(ctx, filter)
	if err != nil {
		log.Error("delete game:%s error:%s", gameID, err.Error())
	}
	log.Info("delete game[%v] success with count:%v", gameID, rst.DeletedCount)
	if err = hdl.server.UpdateMonitor(chain); err != nil {
		log.Error("update monitor lastest error:%s", err.Error())
		return
	}
	return
}
