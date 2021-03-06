package twitterspy

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/cz-theng/czkit-go/log"
)

const (
	baseURL = "https://api.twitter.com/2/search/adaptive.json"
)

var (
	ErrTokenForbid = errors.New("Token is forbidden")
	ErrToken       = errors.New("Get token error")
)

type TwitterSearchConn struct {
	cli       *http.Client
	url       string
	token     string
	tokenRPC  string
	userAgent string
}

func NewTwitterSearchConn() *TwitterSearchConn {
	tsc := &TwitterSearchConn{}
	return tsc
}

func (tsc *TwitterSearchConn) Init(tokenRPC string) (err error) {
	tsc.tokenRPC = tokenRPC
	tsc.cli = &http.Client{}
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents)))); err == nil {
		tsc.userAgent = userAgents[r.Int64()]
	}
	tsc.url = baseURL
	if err = tsc.RefreshToken(); err != nil {
		log.Error("Refresh token error")
		return
	}
	return nil
}

func (tsc *TwitterSearchConn) RefreshToken() (err error) {

	req, err := http.NewRequest("GET", tsc.tokenRPC, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return
	}

	resp, err := tsc.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()

	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Token  string `json:"token"`
		Err    int    `json:"errno"`
		ErrMsg string `json:"errmsg"`
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	if respJOSN.Err > 0 {
		return ErrToken
	}
	tsc.token = respJOSN.Token
	return nil
}

func (tsc *TwitterSearchConn) solidParams() map[string]string {
	parmas := map[string]string{
		"include_can_media_tag":          "1",
		"include_ext_alt_text":           "true",
		"include_quote_count":            "true",
		"include_reply_count":            "1",
		"tweet_mode":                     "extended",
		"include_entities":               "true",
		"include_user_entities":          "true",
		"include_ext_media_availability": "true",
		"send_error_codes":               "true",
		"simple_quoted_tweet":            "true",
		//"count":                          "2",
		"cursor":               "-1",
		"spelling_corrections": "1",
		"ext":                  "mediaStats%2ChighlightedLabel",
		"tweet_search_mode":    "live",
		"f":                    "tweets",
	}
	return parmas
}

//func (tsc *TwitterSearchConn) QueryV(v string, internal time.Duration, count uint32) (tweets map[string]TweetInfo, err error) {
func (tsc *TwitterSearchConn) QueryV(v string, since, until time.Time, count uint32) (tweets map[string]TweetInfo, err error) {
	//log.Info("query for %s ", v)

	params := tsc.solidParams()
	sinceStr := strconv.FormatUint(uint64(since.Unix()), 10)
	untilStr := strconv.FormatUint(uint64(until.Unix()), 10)
	q := "from:" + v
	q += " since:" + sinceStr
	q += " until:" + untilStr

	params["q"] = q
	params["count"] = strconv.Itoa(int(count))

	req, err := http.NewRequest("GET", tsc.url, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return
	}

	qv := req.URL.Query()
	for k, v := range params {
		qv.Add(k, v)
	}
	req.URL.RawQuery = qv.Encode()
	req.Header.Set("User-Agent", tsc.userAgent)
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	//log.Info("token:%s", tsc.token)
	req.Header.Set("x-guest-token", tsc.token)

	resp, err := tsc.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()

	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Errors []struct {
			Code    int
			Message string
		}
		GlobalObjects struct {
			Tweets map[string]TweetInfo
		}
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	//log.Info("resp is :%v", respJOSN)
	if respJOSN.Errors != nil || len(respJOSN.Errors) > 0 {
		return nil, ErrTokenForbid
	}
	tweets = respJOSN.GlobalObjects.Tweets
	return
}

func (tsc *TwitterSearchConn) QueryUserInfo(v string) (user *TwitterUserInfo, err error) {
	apiUrl := `https://twitter.com/i/api/graphql/jMaTS-_Ea8vh9rpKggJbCQ/UserByScreenName?variables=%7B%22screen_name%22%3A%22` + v + `%22%2C%22withHighlightedLabel%22%3Atrue%7D`

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return
	}

	req.Header.Set("User-Agent", tsc.userAgent)
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("x-guest-token", tsc.token)

	resp, err := tsc.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// log.Info("body:%v", string(body))

	// return
	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Errors []struct {
			Code    int
			Message string
		}
		Data struct {
			User TwitterUserInfo `json:"user"`
		} `json:"data"`
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	if respJOSN.Errors != nil || len(respJOSN.Errors) > 0 {
		return nil, ErrTokenForbid
	}
	//log.Info("user:%v", respJOSN.Data.User)
	return &respJOSN.Data.User, nil
}
