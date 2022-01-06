package twitterspy

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/cz-theng/czkit-go/log"
)

const (
	//_twitterIndexURL = "https://twitter.com"
	_twitterIndexURL = "https://api.twitter.com/1.1/guest/activate.json"
)

var (
	ErrNoToken = errors.New("Refresh with no token")
)

type Token struct {
	cli       *http.Client
	token     string
	regexp    *regexp.Regexp
	userAgent string
}

func NewToken() *Token {
	t := &Token{}
	t.cli = &http.Client{}
	t.regexp = regexp.MustCompile(`\("gt=(\d+);`)
	return t
}

func (t *Token) Value() string {
	return t.token
}

func (t *Token) Refresh() error {

	req, err := http.NewRequest("POST", _twitterIndexURL, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return err
	}

	req.Header.Set("User-Agent", UserAgent())
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	resp, err := t.cli.Do(req)
	if err != nil {
		log.Error("cli.Do error:%s", err.Error())
		return err
	}
	defer resp.Body.Close()

	bodyDecoder := json.NewDecoder(resp.Body)
	respJOSN := &struct {
		Code       int    `json:"code"`
		Message    string `json:"message"`
		GuestToken string `json:"guest_token"`
	}{}
	if err = bodyDecoder.Decode(respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return err
	}

	if respJOSN.Code != 0 {
		return errors.New(respJOSN.Message)
	}
	t.token = respJOSN.GuestToken
	return nil
}
