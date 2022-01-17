package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
)

type TweetHandler struct {
	URLHdl
}

//Post is POST
func (hdl *TweetHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	type PostRequest struct {
		TweetID string `json:"tid"`
		Status  int    `json:"status"`
	}
	var req PostRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Error("Decode request error:%s", err.Error())
		encoder.Encode(ErrParam)
		return
	}
	if len(req.TweetID) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	tid, err := strconv.ParseInt(req.TweetID, 10, 64)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}

	if req.Status != int(db.TSDone) &&
		req.Status != int(db.TSFound) {
		encoder.Encode(ErrParam)
		return
	}

	tweetTbl := hdl.server.db.Collection(db.TweetTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"status": req.Status,
		},
	}

	_, err = tweetTbl.UpdateByID(ctx, tid, update, opt)
	if err != nil {
		log.Error("Update vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	type Response struct {
		TweetID string `json:"tid"`
		Status  int    `json:"status"`
		Err     int    `json:"errno"`
		ErrMsg  string `json:"errmsg"`
	}

	rsp := Response{
		TweetID: req.TweetID,
		Status:  req.Status,
		Err:     0,
		ErrMsg:  "",
	}
	encoder.Encode(rsp)
}

//Post is DELETE
func (hdl *TweetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	tweetID := r.FormValue("tid")
	if len(tweetID) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	tid, err := strconv.ParseInt(tweetID, 10, 64)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}

	tweetTbl := hdl.server.db.Collection(db.TweetTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	filter := bson.M{
		"_id": tid,
	}
	_, err = tweetTbl.DeleteMany(ctx, filter)
	if err != nil {
		log.Error("Delete vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	type Response struct {
		TweetID string `json:"tid"`
		Status  int    `json:"status"`
		Err     int    `json:"errno"`
		ErrMsg  string `json:"errmsg"`
	}

	rsp := Response{
		TweetID: tweetID,
		Status:  0,
		Err:     0,
		ErrMsg:  "",
	}
	encoder.Encode(rsp)
}

//Get is GET
func (hdl *TweetHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	tweetTbl := hdl.server.db.Collection(db.TweetTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)
	filter := bson.M{
		"status": db.TSFound,
	}
	cursor, err := tweetTbl.Find(ctx, filter)
	if err != nil {
		log.Error("Find vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	var tweets []db.Tweet
	if err = cursor.All(ctx, &tweets); err != nil {
		log.Error("Find vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	type Response struct {
		Tweets []db.Tweet `json:"tweets"`
		Err    int        `json:"errno"`
		ErrMsg string     `json:"errmsg"`
	}

	rsp := Response{
		Tweets: tweets,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}
