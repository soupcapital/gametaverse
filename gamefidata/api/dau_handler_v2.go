package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/rpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DauHandlerV2 struct {
	URLHdl
}

//Post is POST
func (hdl *DauHandlerV2) Post(w http.ResponseWriter, r *http.Request) {
}

//Delete is DELETE
func (hdl *DauHandlerV2) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *DauHandlerV2) Get(w http.ResponseWriter, r *http.Request) {
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

	theDay := startTS
	type DayInfo struct {
		Date int64 `json:"date"`
		DAU  int   `json:"dau"`
	}

	contracts, err := hdl.server.getContracts(game)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("dauByDate: %s", err.Error())
	}
	var days []DayInfo
	for {
		if theDay > endTS {
			break
		}
		dau, err := hdl.dauByDate(contracts, theDay, theDay+cSecondofDay)
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
	// totalDau, err := hdl.dauByDate(contracts, startTS, endTS+cSecondofDay)
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
		Game:  game,
		Data:  days,
		Total: 0,
	}
	encoder.Encode(rsp)
}

func (hdl *DauHandlerV2) dauByDate(contracts []*pb.Contract, start, end int64) (dau int, err error) {
	conn, err := grpc.Dial(RPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("RPC did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(hdl.server.ctx, 300*time.Second)
	defer cancel()
	dauRsp, err := c.Dau(ctx, &pb.GameReq{
		Start:     start,
		End:       end,
		Contracts: contracts,
	})
	if err != nil {
		log.Error("Dau error:%s", err.Error())
		return
	}
	dau = int(dauRsp.Dau)
	return dau, nil
}
