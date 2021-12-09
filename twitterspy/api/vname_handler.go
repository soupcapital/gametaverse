package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
)

type VNameHandler struct {
	URLHdl
}

//Post is POST
func (hdl *VNameHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vname := r.FormValue("vname")
	if len(vname) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	vnameTbl := hdl.server.db.Collection(db.VNameTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	update := bson.M{
		"$set": bson.M{
			"status": 1,
		},
	}
	_, err := vnameTbl.UpdateByID(ctx, vname, update, opt)
	if err != nil {
		log.Error("Update vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	type Response struct {
		VName  string `json:"vanme"`
		Status int    `json:"status"`
		Err    int    `json:"errno"`
		ErrMsg string `json:"errmsg"`
	}

	rsp := Response{
		VName:  vname,
		Status: 1,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}

//Post is DELETE
func (hdl *VNameHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vname := r.FormValue("vname")

	if len(vname) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	vnameTbl := hdl.server.db.Collection(db.VNameTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	filter := bson.M{
		"_id": vname,
	}
	_, err := vnameTbl.DeleteMany(ctx, filter)
	if err != nil {
		log.Error("Delete vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}
	type Response struct {
		VName  string `json:"vanme"`
		Status int    `json:"status"`
		Err    int    `json:"errno"`
		ErrMsg string `json:"errmsg"`
	}

	rsp := Response{
		VName:  vname,
		Status: 0,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}

//Get is GET
func (hdl *VNameHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vnameTbl := hdl.server.db.Collection(db.VNameTable)

	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	opt := mngopts.Update()
	opt.SetUpsert(true)

	cursor, err := vnameTbl.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Find vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	var vnames []db.VName
	if err = cursor.All(ctx, &vnames); err != nil {
		log.Error("Find vname error: ", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	type Response struct {
		VNames []string `json:"vanmes"`
		Err    int      `json:"errno"`
		ErrMsg string   `json:"errmsg"`
	}

	var names []string
	for _, vname := range vnames {
		names = append(names, vname.ID)
	}

	rsp := Response{
		VNames: names,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}
