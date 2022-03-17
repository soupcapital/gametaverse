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

type AllHandlerV2 struct {
	URLHdl
}

//Delete is DELETE
func (hdl *AllHandlerV2) Delete(w http.ResponseWriter, r *http.Request) {
}

//Post is POST
func (hdl *AllHandlerV2) Post(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *AllHandlerV2) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with sort")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	start := r.FormValue("start")
	end := r.FormValue("end")

	if len(end) == 0 ||
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

	dau, err := hdl.dauByDate(startTS, endTS+cSecondofDay)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("daus: %s", err.Error())
		return
	}
	trxes, err := hdl.trxByDate(startTS, endTS+cSecondofDay)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("trxes: %s", err.Error())
		return
	}

	type Response struct {
		DAU int `json:"dau"`
		Trx int `json:"trx"`
	}

	rsp := Response{
		DAU: dau,
		Trx: trxes,
	}
	encoder.Encode(rsp)
}

func (hdl *AllHandlerV2) dauByDate(start, end int64) (dau int, err error) {
	conn, err := grpc.Dial(RPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("RPC did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(hdl.server.ctx, 3*time.Second)
	defer cancel()
	dauRsp, err := c.ChainDau(ctx, &pb.ChainGameReq{
		Start:  start,
		End:    end,
		Chains: []pb.Chain{pb.Chain_BSC, pb.Chain_ETH, pb.Chain_POLYGON},
	})
	if err != nil {
		log.Error("Dau error:%s", err.Error())
		return
	}
	dau = int(dauRsp.Dau)
	return dau, nil
}

func (hdl *AllHandlerV2) trxByDate(start, end int64) (count int, err error) {
	conn, err := grpc.Dial(RPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("RPC did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(hdl.server.ctx, 3*time.Second)
	defer cancel()
	countRsp, err := c.ChainTxCount(ctx, &pb.ChainGameReq{
		Start:  start,
		End:    end,
		Chains: []pb.Chain{pb.Chain_BSC, pb.Chain_ETH, pb.Chain_POLYGON},
	})
	if err != nil {
		log.Error("Dau error:%s", err.Error())
		return
	}
	count = int(countRsp.Count)
	return count, nil
}
