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

type User2GameHandler struct {
	URLHdl
}

//Post is POST
func (hdl *User2GameHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Delete is DELETE
func (hdl *User2GameHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *User2GameHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("deal with dau")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	start := r.FormValue("start")
	end := r.FormValue("end")
	user := r.FormValue("user")

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

	type Response struct {
		Games []string `json:"games"`
	}
	var users []*pb.Contract

	for _, c := range AllChain {
		users = append(users, &pb.Contract{
			Chain:   c,
			Address: user,
		})
	}

	contracts, err := hdl.getUserContracts(users, startTS, endTS)
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("trxes: %s", err.Error())
		return
	}

	games, err := hdl.server.getAllGames()
	if err != nil {
		encoder.Encode(ErrDB)
		log.Error("get all games: %s", err.Error())
		return
	}

	var gameCount map[string]int = make(map[string]int, len(games))

	for _, c := range contracts {
		for _, g := range games {
			for _, gc := range g.Contracts {
				if c == gc {
					gameCount[g.GameID] += 1
				}
			}
		}
	}
	rsp := Response{}
	for k := range gameCount {
		rsp.Games = append(rsp.Games, k)
	}
	encoder.Encode(rsp)
}

func (hdl *User2GameHandler) getUserContracts(users []*pb.Contract, start, end int64) (contracts []string, err error) {
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
	req := &pb.AllUserProgramsReq{
		Start: start,
		End:   end,
		Users: users,
	}
	rsp, err := c.AllUserPrograms(ctx, req)
	if err != nil {
		log.Error("Dau error:%s", err.Error())
		return
	}
	contracts = make([]string, len(rsp.Programs))
	copy(contracts, rsp.Programs[:])
	return
}
