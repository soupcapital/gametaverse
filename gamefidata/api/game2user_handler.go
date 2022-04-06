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

type Game2UserHandler struct {
	URLHdl
}

//Post is POST
func (hdl *Game2UserHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Delete is DELETE
func (hdl *Game2UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *Game2UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	start := r.FormValue("start")
	end := r.FormValue("end")
	gameOneID := r.FormValue("game1")
	gameTwoID := r.FormValue("game2")

	if len(end) == 0 ||
		len(start) == 0 ||
		len(gameOneID) == 0 ||
		len(gameTwoID) == 0 {
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

	gameOne, err := hdl.server.getGameContracts(gameOneID)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("get gameOne: %s", err.Error())
		return
	}

	gameTwo, err := hdl.server.getGameContracts(gameTwoID)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("get gameOne: %s", err.Error())
		return
	}

	users, err := hdl.getGameUsers(gameOne, gameTwo, startTS, endTS)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("trxes: %s", err.Error())
		return
	}

	type Response struct {
		Users []string `json:"users"`
	}
	rsp := Response{
		Users: users,
	}
	encoder.Encode(rsp)
}

func (hdl *Game2UserHandler) getGameUsers(gameOne []*pb.Contract, gameTwo []*pb.Contract, start, end int64) (users []string, err error) {
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
	req := &pb.TwoGamesPlayersReq{
		Start:   start,
		End:     end,
		GameOne: gameOne,
		GameTwo: gameTwo,
	}
	rsp, err := c.TwoGamesPlayers(ctx, req)
	if err != nil {
		log.Error("Dau error:%s", err.Error())
		return
	}
	users = make([]string, len(rsp.Users))
	copy(users, rsp.Users[:])
	return
}
