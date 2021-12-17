package api

import (
	"encoding/json"
	"net/http"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
)

type UserInfoHandler struct {
	URLHdl
}

//Post is POST
func (hdl *UserInfoHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Post is DELETE
func (hdl *UserInfoHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *UserInfoHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vname := r.FormValue("vname")
	if len(vname) == 0 {
		encoder.Encode(ErrParam)
		return
	}

	if len(hdl.server.token.Value()) == 0 {
		if err := hdl.server.token.Refresh(); err != nil {
			log.Error("Refresh token error:%s", err.Error())
			encoder.Encode(ErrParam)
			return
		}
	}

	userInfo, err := hdl.server.conn.QueryUserInfo(vname)

	if err == twitterspy.ErrTokenForbid {
		if err = hdl.server.token.Refresh(); err != nil {
			log.Error("Refresh token error:%s", err.Error())
			encoder.Encode(ErrParam)
			return
		}
		userInfo, err = hdl.server.conn.QueryUserInfo(vname)
		if err != nil {
			log.Error("QueryUserInfo  error:%s", err.Error())
			encoder.Encode(ErrUserInfo)
			return
		}
	}

	type Response struct {
		User   *twitterspy.TwitterUserInfo `json:"user"`
		Err    int                         `json:"errno"`
		ErrMsg string                      `json:"errmsg"`
	}

	rsp := Response{
		User:   userInfo,
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}
