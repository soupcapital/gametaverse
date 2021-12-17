package twitterspy

import (
	"bytes"
	"net/http"
	"regexp"

	"github.com/cz-theng/czkit-go/log"
)

const (
	_twitterIndexURL = "https://twitter.com"
	_bearer          = `Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA`
)

type Token struct {
	cli    *http.Client
	token  string
	regexp *regexp.Regexp
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
	req, err := http.NewRequest("GET", _twitterIndexURL, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return err
	}

	resp, err := t.cli.Do(req)
	if err != nil {
		log.Error("cli.Do error:%s", err.Error())
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		log.Error("ReadFrom response  error:%s", err.Error())
		return err
	}

	tBuf := t.regexp.Find(buf.Bytes())
	if tBuf == nil {
		log.Info("no token")
	} else {
		t.token = string(tBuf[5 : len(tBuf)-1])
		log.Info("token is %s", t.token)
	}

	return nil
}
