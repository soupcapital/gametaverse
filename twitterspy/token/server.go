package token

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
)

type Server struct {
	ctx       context.Context
	cancelFun context.CancelFunc
	opts      options
	httpd     http.Server
	router    *Router
	token     *twitterspy.Token
}

func NewServer() (svr *Server) {
	svr = &Server{}
	return svr
}

func (svr *Server) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&svr.opts)
	}

	svr.ctx, svr.cancelFun = context.WithCancel(context.Background())

	svr.httpd = http.Server{
		Addr:           svr.opts.ListenAddr,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        svr,
	}
	svr.router = NewRouter()
	svr.token = twitterspy.NewToken()
	svr.initHandler()
	return
}

func (svr *Server) initHandler() {
	svr.router.RegistRaw("/twitterspy/api/v1/token", &TokenHandler{URLHdl{server: svr}})
}

func (svr *Server) Run() (err error) {

	//go svr.tokenLoop()

	err = svr.httpd.ListenAndServe()
	if err != nil {
		log.Error("ListenAndServe Error:%s", err.Error())
	}
	return
}

func (svr *Server) tokenLoop() {
	svr.queryToken()
	ticker := time.NewTicker(6 * 60 * 60 * time.Second)
	for range ticker.C {
		svr.queryToken()
	}
}

func (svr *Server) queryToken() {
	for {
		if err := svr.token.Refresh(); err == nil {
			log.Info("refresh token success and got :%s", svr.token.Value())
			return
		} else {
			log.Error("refresh toekn faild:%s", err.Error())
		}
		time.Sleep(7 * time.Second)
	}
}

func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svr.router.DealRaw(r.URL.Path, w, r)
}

func (svr *Server) checkToken() (valied bool) {
	v := "ForM214"
	apiUrl := `https://twitter.com/i/api/graphql/jMaTS-_Ea8vh9rpKggJbCQ/UserByScreenName?variables=%7B%22screen_name%22%3A%22` + v + `%22%2C%22withHighlightedLabel%22%3Atrue%7D`

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return false
	}

	userAgent := twitterspy.UserAgent()

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("x-guest-token", svr.token.Value())
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return false
	}
	defer resp.Body.Close()

	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Errors []struct {
			Code    int
			Message string
		}
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	if respJOSN.Errors != nil || len(respJOSN.Errors) > 0 {
		return false
	}
	return true
}
