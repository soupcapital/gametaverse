package twitterspy

import (
	"bytes"
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"regexp"

	"github.com/cz-theng/czkit-go/log"
)

const (
	_twitterIndexURL = "https://twitter.com"
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
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents)))); err == nil {
		t.userAgent = userAgents[r.Int64()]
	}
	t.regexp = regexp.MustCompile(`\("gt=(\d+);`)
	return t
}

func (t *Token) Value() string {
	return t.token
}

func (t *Token) Refresh() error {
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents)))); err == nil {
		t.userAgent = userAgents[r.Int64()]
	}
	req, err := http.NewRequest("GET", _twitterIndexURL, nil)
	if err != nil {
		log.Error("create request error:%s", err.Error())
		return err
	}

	//req.Header.Set("User-Agent", t.userAgent)
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
	//log.Info("buf:%s", string(buf.Bytes()))

	tBuf := t.regexp.Find(buf.Bytes())
	if tBuf == nil {
		log.Info("no token")
		return ErrNoToken
	} else {
		t.token = string(tBuf[5 : len(tBuf)-1])
		log.Info("token is %s", t.token)
	}

	return nil
}
