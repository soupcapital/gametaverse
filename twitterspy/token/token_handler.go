package token

import (
	"encoding/json"
	"net/http"
)

type TokenHandler struct {
	URLHdl
}

//Post is POST
func (hdl *TokenHandler) Post(w http.ResponseWriter, r *http.Request) {
}

//Post is DELETE
func (hdl *TokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

//Get is GET
func (hdl *TokenHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	if !hdl.server.checkToken() {
		hdl.server.queryToken()
	}

	type Response struct {
		Token  string `json:"token"`
		Err    int    `json:"errno"`
		ErrMsg string `json:"errmsg"`
	}

	rsp := Response{
		Token:  hdl.server.token.Value(),
		Err:    0,
		ErrMsg: "",
	}
	encoder.Encode(rsp)
}
