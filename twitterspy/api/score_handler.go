package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
)

type ScoreHandler struct {
	URLHdl
}

//Post is POST
func (hdl *ScoreHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Post is DELETE
func (hdl *ScoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *ScoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	//vname := r.FormValue("vname")
	date := r.FormValue("date")

	if len(date) == 0 {
		log.Error(" date:%s", date)
		encoder.Encode(ErrParam)
		return
	}

	dayTime, err := time.Parse(twitterspy.DateFormat, date)
	if err != nil {
		encoder.Encode(ErrParam)
		return
	}
	dayTS := dayTime.Unix()

	if dayTS%twitterspy.SecOfDay != 0 {
		log.Error("dateTS:%v", dayTS)
		encoder.Encode(ErrTimestamp)
		return
	}

	if dayTS <= twitterspy.SecOfDay {
		encoder.Encode(ErrTimestamp)
		return
	}

	dayInfos, err := hdl.queryDiggerInfosForOneDay(dayTS)
	if err != nil {
		log.Error("queryDiggerInfoForOneDay day info error:%s", err.Error())
		encoder.Encode(ErrDB)
		return
	}

	sort.Slice(dayInfos, func(i, j int) bool {
		return dayInfos[i].FavoriteCount >= dayInfos[j].FavoriteCount
	})

	type BillboardItem struct {
		Name  string  `json:"name"`
		Score float32 `json:"score"`
	}

	type Response struct {
		Billboard []*BillboardItem
		Err       int    `json:"errno"`
		ErrMsg    string `json:"errmsg"`
	}
	var items []*BillboardItem
	maxLen := 100
	if maxLen > len(dayInfos) {
		maxLen = len(dayInfos)
	}
	for i := 0; i < maxLen; i++ {
		item := &BillboardItem{
			Name:  dayInfos[i].Name,
			Score: dayInfos[i].Score,
		}
		items = append(items, item)
	}

	rsp := Response{
		Billboard: items,
		Err:       0,
		ErrMsg:    "",
	}
	encoder.Encode(rsp)
}

func (hdl *ScoreHandler) queryDiggerInfosForOneDay(dateTS int64) (infos []*db.Digger, err error) {
	ctx, cancel := context.WithTimeout(hdl.server.ctx, 1000*time.Second)
	defer cancel()

	diggerTbl := hdl.server.db.Collection(db.DiggerTable)
	curs, err := diggerTbl.Find(ctx, bson.M{"ts": dateTS})
	if err != nil {
		log.Error("Find ts[%v] error", dateTS)
		return
	}
	if err = curs.All(ctx, &infos); err != nil {
		log.Error("Decode digger infos error: ", err.Error())
		return
	}
	return
}
